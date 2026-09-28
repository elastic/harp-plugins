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
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ResolveGitContext resolves git provenance for the given spec file path.
// It returns the repo name (e.g. "elastic/harp-plugins"), the repo-relative
// path to the file, and the HEAD commit hash. The commit hash is suffixed
// with "+dirty" when the spec file has uncommitted local changes.
//
// Returns an error when the path is not inside a git repository.
func ResolveGitContext(specPath string) (gitRepo, sourceFile, gitCommit string, err error) {
	absPath, err := filepath.Abs(specPath)
	if err != nil {
		return "", "", "", fmt.Errorf("unable to resolve spec path: %w", err)
	}

	specDir := filepath.Dir(absPath)

	// Locate the git root.
	rootOut, err := gitOutput(specDir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", "", "", fmt.Errorf("unable to determine git root (is the path inside a git repo?): %w", err)
	}
	gitRoot := rootOut

	// Compute the repo-relative path, normalised to forward slashes.
	relPath, err := filepath.Rel(gitRoot, absPath)
	if err != nil {
		return "", "", "", fmt.Errorf("unable to compute relative path: %w", err)
	}
	sourceFile = filepath.ToSlash(relPath)

	// Derive the repo name from the remote URL; fall back to the root dir name.
	remoteOut, remoteErr := gitOutput(specDir, "remote", "get-url", "origin")
	if remoteErr == nil {
		gitRepo = parseRepoFromRemoteURL(remoteOut)
	} else {
		gitRepo = filepath.Base(gitRoot)
	}

	// Resolve HEAD commit hash.
	commitOut, err := gitOutput(specDir, "rev-parse", "HEAD")
	if err != nil {
		return "", "", "", fmt.Errorf("unable to determine git commit: %w", err)
	}
	gitCommit = commitOut

	// Append +dirty when the spec file has local uncommitted changes.
	statusOut, statusErr := gitOutput(specDir, "status", "--porcelain", absPath)
	if statusErr == nil && statusOut != "" {
		gitCommit += "+dirty"
	}

	return gitRepo, sourceFile, gitCommit, nil
}

// parseRepoFromRemoteURL extracts the "org/repo" segment from a git remote URL.
// Supports HTTPS (https://github.com/org/repo.git) and SSH (git@github.com:org/repo.git) forms.
func parseRepoFromRemoteURL(remoteURL string) string {
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	// SSH form: git@github.com:org/repo  — the colon is not followed by "//"
	if idx := strings.LastIndex(remoteURL, ":"); idx >= 0 {
		candidate := remoteURL[idx+1:]
		if strings.Contains(candidate, "/") && !strings.HasPrefix(candidate, "//") {
			return candidate
		}
	}

	// HTTPS form: https://host/org/repo — take the last two path segments.
	parts := strings.Split(remoteURL, "/")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], "/")
	}

	return remoteURL
}

// gitOutput runs a git command inside dir and returns trimmed stdout.
func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
