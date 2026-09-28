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

// ResolveGitContext resolves git provenance for the given spec file path.
// It returns the repo name (e.g. "elastic/harp-plugins"), the repo-relative
// path to the file, and the HEAD commit hash. The commit hash is suffixed
// with "+dirty" when the spec file has uncommitted local changes.
//
// Returns an error when the path is not inside a git repository.
func ResolveGitContext(ctx context.Context, specPath string) (gitRepo, sourceFile, gitCommit string, err error) {
	absPath, err := filepath.Abs(specPath)
	if err != nil {
		return "", "", "", fmt.Errorf("unable to resolve spec path: %w", err)
	}

	specDir := filepath.Dir(absPath)

	// Locate the git root.
	gitRoot, err := gitOutput(ctx, specDir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", "", "", fmt.Errorf("unable to determine git root (is the path inside a git repo?): %w", err)
	}

	// Compute the repo-relative path, normalised to forward slashes.
	relPath, err := filepath.Rel(gitRoot, absPath)
	if err != nil {
		return "", "", "", fmt.Errorf("unable to compute relative path: %w", err)
	}
	sourceFile = sanitizeGitString(filepath.ToSlash(relPath))

	// Derive the repo name from the remote URL; fall back to the root dir name.
	remoteOut, remoteErr := gitOutput(ctx, specDir, "remote", "get-url", "origin")
	if remoteErr == nil {
		gitRepo = sanitizeGitString(parseRepoFromRemoteURL(remoteOut))
	} else {
		gitRepo = sanitizeGitString(filepath.Base(gitRoot))
	}

	// Resolve HEAD commit hash.
	commitOut, err := gitOutput(ctx, specDir, "rev-parse", "HEAD")
	if err != nil {
		return "", "", "", fmt.Errorf("unable to determine git commit: %w", err)
	}
	gitCommit = sanitizeGitString(commitOut)

	// Append +dirty when the spec file has local uncommitted changes.
	statusOut, statusErr := gitOutput(ctx, specDir, "status", "--porcelain", absPath)
	if statusErr == nil && statusOut != "" {
		gitCommit += "+dirty"
	}

	return gitRepo, sourceFile, gitCommit, nil
}

// parseRepoFromRemoteURL extracts the "org/repo" segment from a git remote URL.
// Supports HTTPS (https://github.com/org/repo.git) and SSH (git@github.com:org/repo.git) forms.
// SSH URLs with an explicit port (ssh://git@host:22/org/repo) fall through to the HTTPS path.
func parseRepoFromRemoteURL(remoteURL string) string {
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	// SSH form: git@github.com:org/repo  — colon not followed by "//" and not a port number.
	if idx := strings.LastIndex(remoteURL, ":"); idx >= 0 {
		candidate := remoteURL[idx+1:]
		if strings.Contains(candidate, "/") && !strings.HasPrefix(candidate, "//") {
			// Reject port numbers: the segment before the first "/" must not be all digits.
			beforeSlash := candidate[:strings.Index(candidate, "/")]
			isPort := len(beforeSlash) > 0
			for _, r := range beforeSlash {
				if !unicode.IsDigit(r) {
					isPort = false
					break
				}
			}
			if !isPort {
				return candidate
			}
		}
	}

	// HTTPS (or ssh:// with port) form: take the last two path segments.
	parts := strings.Split(remoteURL, "/")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], "/")
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

// sanitizeGitString removes embedded newline characters from git output before
// it is embedded in HCL comment lines, preventing comment-injection attacks.
func sanitizeGitString(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, s)
}
