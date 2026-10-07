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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Test_parseRepoFromRemoteURL(t *testing.T) {
	tests := []struct {
		name      string
		remoteURL string
		want      string
	}{
		{
			name:      "https with .git suffix",
			remoteURL: "https://github.com/elastic/harp-plugins.git",
			want:      "elastic/harp-plugins",
		},
		{
			name:      "https without .git suffix",
			remoteURL: "https://github.com/elastic/harp-plugins",
			want:      "elastic/harp-plugins",
		},
		{
			name:      "ssh with .git suffix",
			remoteURL: "git@github.com:elastic/harp-plugins.git",
			want:      "elastic/harp-plugins",
		},
		{
			name:      "ssh without .git suffix",
			remoteURL: "git@github.com:elastic/harp-plugins",
			want:      "elastic/harp-plugins",
		},
		{
			name:      "gitlab ssh",
			remoteURL: "git@gitlab.com:myorg/myrepo.git",
			want:      "myorg/myrepo",
		},
		{
			name:      "https subdomain",
			remoteURL: "https://gitlab.example.com/myorg/myrepo.git",
			want:      "myorg/myrepo",
		},
		{
			name:      "ssh with explicit port",
			remoteURL: "ssh://git@github.com:22/elastic/harp-plugins.git",
			want:      "elastic/harp-plugins",
		},
		{
			name:      "gitlab https nested subgroup",
			remoteURL: "https://gitlab.com/group/subgroup/repo.git",
			want:      "group/subgroup/repo",
		},
		{
			name:      "gitlab ssh nested subgroup",
			remoteURL: "git@gitlab.com:group/subgroup/repo.git",
			want:      "group/subgroup/repo",
		},
		{
			name:      "ssh url nested subgroup with port",
			remoteURL: "ssh://git@gitlab.com:2222/group/subgroup/repo.git",
			want:      "group/subgroup/repo",
		},
		{
			name:      "https credentials stripped",
			remoteURL: "https://user:tok@github.com/elastic/harp-plugins.git",
			want:      "elastic/harp-plugins",
		},
		{
			name:      "https token-only userinfo stripped",
			remoteURL: "https://tok@github.com/elastic/harp-plugins.git",
			want:      "elastic/harp-plugins",
		},
		{
			name:      "https trailing slash",
			remoteURL: "https://github.com/elastic/harp-plugins/",
			want:      "elastic/harp-plugins",
		},
		{
			name:      "https .git with trailing slash",
			remoteURL: "https://github.com/elastic/harp-plugins.git/",
			want:      "elastic/harp-plugins",
		},
		{
			name:      "ssh numeric org",
			remoteURL: "git@github.com:123/repo.git",
			want:      "123/repo",
		},
		{
			name:      "surrounding whitespace",
			remoteURL: "  https://github.com/elastic/harp-plugins.git\n",
			want:      "elastic/harp-plugins",
		},
		{
			name:      "empty",
			remoteURL: "",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRepoFromRemoteURL(tt.remoteURL)
			if got != tt.want {
				t.Errorf("parseRepoFromRemoteURL(%q) = %q, want %q", tt.remoteURL, got, tt.want)
			}
		})
	}
}

// hermeticGit prepares the environment so git behaves identically on every
// machine: no user or system config, and no ambient repository selection.
// It skips the test when git is not installed. Tests using it must not run in
// parallel because they modify process environment.
func hermeticGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, k := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE"} {
		t.Setenv(k, "") // registers restore of the original value
		if err := os.Unsetenv(k); err != nil {
			t.Fatal(err)
		}
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

// newRepo creates an initialized repo containing specs/a.yaml and returns the
// repo root and the spec path.
func newRepo(t *testing.T, dir string) (root, spec string) {
	t.Helper()
	runGit(t, dir, "init", "-q")
	spec = filepath.Join(dir, "specs", "a.yaml")
	if err := os.MkdirAll(filepath.Dir(spec), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(spec, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir, spec
}

func Test_ResolveGitContext_nonGitPath(t *testing.T) {
	hermeticGit(t)
	tmpDir := t.TempDir()
	fakeSpec := filepath.Join(tmpDir, "spec.yaml")
	if err := os.WriteFile(fakeSpec, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := ResolveGitContext(context.Background(), fakeSpec); err == nil {
		t.Error("expected error for path not in a git repo, got nil")
	}
}

func Test_ResolveGitContext_noOrigin(t *testing.T) {
	hermeticGit(t)
	_, spec := newRepo(t, t.TempDir())

	got, err := ResolveGitContext(context.Background(), spec)
	if err != nil {
		t.Fatalf("ResolveGitContext() error = %v", err)
	}
	if got.GitRepo != "" {
		t.Errorf("GitRepo = %q, want empty when no origin remote is configured", got.GitRepo)
	}
	if got.SourceFile != "specs/a.yaml" {
		t.Errorf("SourceFile = %q, want specs/a.yaml", got.SourceFile)
	}
}

func Test_ResolveGitContext_withOrigin(t *testing.T) {
	hermeticGit(t)
	root, spec := newRepo(t, t.TempDir())
	runGit(t, root, "remote", "add", "origin", "https://user:tok@github.com/elastic/harp-plugins.git")

	got, err := ResolveGitContext(context.Background(), spec)
	if err != nil {
		t.Fatalf("ResolveGitContext() error = %v", err)
	}
	if got.GitRepo != "elastic/harp-plugins" {
		t.Errorf("GitRepo = %q, want elastic/harp-plugins (credentials must not leak)", got.GitRepo)
	}
}

func Test_ResolveGitContext_symlinks(t *testing.T) {
	hermeticGit(t)
	root, spec := newRepo(t, t.TempDir())

	linkDir := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(root, linkDir); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	linkedSpec := filepath.Join(linkDir, "specs", "a.yaml")

	fileLink := filepath.Join(t.TempDir(), "a-link.yaml")
	if err := os.Symlink(spec, fileLink); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	tests := []struct {
		name string
		path string
	}{
		{name: "symlinked directory", path: linkedSpec},
		{name: "symlinked file", path: fileLink},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveGitContext(context.Background(), tt.path)
			if err != nil {
				t.Fatalf("ResolveGitContext() error = %v", err)
			}
			if got.SourceFile != "specs/a.yaml" {
				t.Errorf("SourceFile = %q, want specs/a.yaml", got.SourceFile)
			}
		})
	}
}

func Test_ResolveGitContext_relativePath(t *testing.T) {
	hermeticGit(t)
	root, _ := newRepo(t, t.TempDir())
	t.Chdir(filepath.Join(root, "specs"))

	got, err := ResolveGitContext(context.Background(), "a.yaml")
	if err != nil {
		t.Fatalf("ResolveGitContext() error = %v", err)
	}
	if got.SourceFile != "specs/a.yaml" {
		t.Errorf("SourceFile = %q, want specs/a.yaml", got.SourceFile)
	}
}

func Test_ResolveGitContext_symlinkEscapesRepo(t *testing.T) {
	hermeticGit(t)
	root, _ := newRepo(t, t.TempDir())

	outside := filepath.Join(t.TempDir(), "outside.yaml")
	if err := os.WriteFile(outside, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	escape := filepath.Join(root, "specs", "escape.yaml")
	if err := os.Symlink(outside, escape); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	got, err := ResolveGitContext(context.Background(), escape)
	if err == nil {
		t.Fatalf("expected error for spec resolving outside the git root, got %+v", got)
	}
	if !strings.Contains(err.Error(), "outside") && !strings.Contains(err.Error(), "git root") {
		t.Errorf("error = %v, want mention of path outside git root", err)
	}
}

func Test_sanitizeGitString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "clean string unchanged",
			input: "elastic/harp-plugins",
			want:  "elastic/harp-plugins",
		},
		{
			name:  "embedded newline removed",
			input: "elastic/harp-plugins\ninjected line",
			want:  "elastic/harp-pluginsinjected line",
		},
		{
			name:  "embedded carriage return removed",
			input: "abc123\rmalicious",
			want:  "abc123malicious",
		},
		{
			name:  "CRLF removed",
			input: "org/repo\r\nother",
			want:  "org/repoother",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "double quote removed",
			input: `a"b`,
			want:  "ab",
		},
		{
			name:  "backslash removed",
			input: `a\b`,
			want:  "ab",
		},
		{
			name:  "tab removed",
			input: "a\tb",
			want:  "ab",
		},
		{
			name:  "NUL removed",
			input: "a\x00b",
			want:  "ab",
		},
		{
			name:  "ANSI escape introducer removed",
			input: "a\x1b[31mb",
			want:  "a[31mb",
		},
		{
			name:  "line and paragraph separators removed",
			input: "a\u2028b\u2029c",
			want:  "abc",
		},
		{
			name:  "unicode letters preserved",
			input: "org/répo-日本",
			want:  "org/répo-日本",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeGitString(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeGitString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
