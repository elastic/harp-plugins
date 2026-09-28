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
	"os"
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

func Test_ResolveGitContext_nonGitPath(t *testing.T) {
	tmpDir := t.TempDir()
	fakeSpec := filepath.Join(tmpDir, "spec.yaml")
	if err := os.WriteFile(fakeSpec, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, _, _, err := ResolveGitContext(fakeSpec)
	if err == nil {
		t.Error("expected error for path not in a git repo, got nil")
	}
}

func Test_ResolveGitContext_withRealRepo(t *testing.T) {
	// compiler.go is a stable committed file in this package directory.
	_, sourceFile, gitCommit, err := ResolveGitContext("compiler.go")
	if err != nil {
		t.Fatalf("ResolveGitContext() error = %v", err)
	}

	if sourceFile == "" {
		t.Error("sourceFile should not be empty")
	}
	if gitCommit == "" {
		t.Error("gitCommit should not be empty")
	}

	if !strings.HasSuffix(sourceFile, "compiler.go") {
		t.Errorf("sourceFile %q should end with compiler.go", sourceFile)
	}
	if strings.Contains(sourceFile, "\\") {
		t.Errorf("sourceFile %q must use forward slashes", sourceFile)
	}

	commit := strings.TrimSuffix(gitCommit, "+dirty")
	if len(commit) == 0 {
		t.Error("gitCommit (without +dirty) must not be empty")
	}
	for _, c := range commit {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("gitCommit %q contains non-hex character %q", commit, string(c))
			break
		}
	}
}

func Test_ResolveGitContext_returnsGitRepo(t *testing.T) {
	gitRepo, _, _, err := ResolveGitContext("compiler.go")
	if err != nil {
		t.Fatalf("ResolveGitContext() error = %v", err)
	}
	if gitRepo == "" {
		t.Error("gitRepo should not be empty")
	}
	parts := strings.Split(gitRepo, "/")
	if len(parts) != 2 {
		t.Errorf("gitRepo %q should be in org/repo form", gitRepo)
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
