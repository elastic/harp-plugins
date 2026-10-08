// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package terraformer

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

var (
	// ErrGitNotFound is returned when the git binary is not available on PATH.
	ErrGitNotFound = errors.New("git binary not found in PATH")
	// ErrNotGitRepo is returned when the spec path is not inside a git repository.
	ErrNotGitRepo = errors.New("path is not inside a git repository")
	// ErrOutsideGitRoot is returned when the spec path resolves outside the
	// git root, for example through a symlink.
	ErrOutsideGitRoot = errors.New("spec path is outside the git root")
)

// ResolveGitContext resolves source provenance for the given spec file path.
// The returned SourceInfo carries the repo name (e.g. "elastic/harp-plugins")
// and the repo-relative path to the file. GitRepo is empty when the repository
// has no origin remote.
//
// Returns an error when the path is not inside a git repository, resolves
// outside of it, or git is not installed (see ErrGitNotFound).
func ResolveGitContext(ctx context.Context, specPath string) (SourceInfo, error) {
	absPath, err := filepath.Abs(specPath)
	if err != nil {
		return SourceInfo{}, fmt.Errorf("unable to resolve spec path: %w", err)
	}

	// git reports the symlink-resolved root, so resolve the spec path the same
	// way to keep the repo-relative path correct (e.g. /var -> /private/var).
	if resolved, evalErr := filepath.EvalSymlinks(absPath); evalErr == nil {
		absPath = resolved
	}

	specDir := filepath.Dir(absPath)

	// Locate the git root.
	gitRoot, err := gitOutput(ctx, specDir, "rev-parse", "--show-toplevel")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return SourceInfo{}, fmt.Errorf("%w: %w", ErrGitNotFound, err)
		}
		return SourceInfo{}, fmt.Errorf("%w: %w", ErrNotGitRepo, err)
	}

	// Compute the repo-relative path, normalised to forward slashes.
	relPath, err := repoRelativePath(gitRoot, absPath)
	if err != nil {
		return SourceInfo{}, err
	}
	src := SourceInfo{SourceFile: sanitizeGitString(relPath)}

	// Derive the repo name from the remote URL. Without an origin remote the
	// name is left empty rather than guessed from the local directory name.
	if remoteOut, remoteErr := gitOutput(ctx, specDir, "remote", "get-url", "origin"); remoteErr == nil {
		src.GitRepo = sanitizeGitString(parseRepoFromRemoteURL(remoteOut))
	}

	return src, nil
}

// repoRelativePath returns absPath relative to gitRoot using forward slashes.
// It fails with ErrOutsideGitRoot when absPath is not under gitRoot.
func repoRelativePath(gitRoot, absPath string) (string, error) {
	relPath, err := filepath.Rel(gitRoot, absPath)
	if err != nil {
		return "", fmt.Errorf("unable to compute relative path: %w", err)
	}
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: %q is not under %q", ErrOutsideGitRoot, absPath, gitRoot)
	}
	return filepath.ToSlash(relPath), nil
}

// parseRepoFromRemoteURL extracts the repository path ("org/repo", or
// "group/subgroup/repo" for nested namespaces) from a git remote URL.
// Supports URL forms (https://host/org/repo.git, ssh://git@host:22/org/repo)
// and scp-like SSH (git@host:org/repo.git).
func parseRepoFromRemoteURL(remoteURL string) string {
	remoteURL = strings.TrimSuffix(strings.TrimRight(strings.TrimSpace(remoteURL), "/"), ".git")

	// URL form: everything after the host is the repo path.
	if _, rest, ok := strings.Cut(remoteURL, "://"); ok {
		if _, path, found := strings.Cut(rest, "/"); found {
			return strings.Trim(path, "/")
		}
		return remoteURL
	}

	// scp-like form: host:path
	if _, path, ok := strings.Cut(remoteURL, ":"); ok {
		return strings.Trim(path, "/")
	}

	return remoteURL
}

// gitOutput runs a git command inside dir and returns trimmed stdout.
// When the command fails, stderr from git is included in the returned error.
func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("gitOutput: at least one git argument required")
	}
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return "", fmt.Errorf("git %s: %s: %w", args[0], strings.TrimSpace(string(exitErr.Stderr)), err)
		}
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}
	return strings.TrimSpace(string(out)), nil
}

// SanitizeSourceValue makes a caller-supplied provenance value safe to embed
// in the generated header, using the same rules as values read from git.
func SanitizeSourceValue(s string) string {
	return sanitizeGitString(s)
}

// sanitizeGitString removes characters from git output that would corrupt the
// quoted HCL comment values it is embedded in: control characters (including
// newlines, NUL and ANSI escapes), the Unicode line and paragraph separators,
// and the quote and backslash characters that would break the surrounding
// string.
func sanitizeGitString(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case unicode.IsControl(r), r == '\u2028', r == '\u2029', r == '"', r == '\\':
			return -1
		}
		return r
	}, s)
}
