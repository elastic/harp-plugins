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

package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/cobra"

	"github.com/elastic/harp-plugins/terraformer/pkg/terraformer"
)

type gitResult struct {
	src terraformer.SourceInfo
	err error
}

func Test_resolveSourceInfo(t *testing.T) {
	gitOK := gitResult{src: terraformer.SourceInfo{GitRepo: "git/repo", SourceFile: "specs/a.yaml"}}

	tests := []struct {
		name       string
		specPath   string
		flags      provenanceFlags
		git        gitResult
		want       terraformer.SourceInfo
		wantErr    error
		wantCalled bool
	}{
		{
			name:     "stdin without overrides yields zero value and no error",
			specPath: "-",
			git:      gitOK,
			want:     terraformer.SourceInfo{},
		},
		{
			name:     "stdin with both overrides uses them and skips git",
			specPath: "-",
			flags:    provenanceFlags{sourceRepo: "elastic/harp-plugins", sourceFile: "specs/a.yaml"},
			git:      gitOK,
			want:     terraformer.SourceInfo{GitRepo: "elastic/harp-plugins", SourceFile: "specs/a.yaml"},
		},
		{
			name:     "overrides are sanitized",
			specPath: "-",
			flags:    provenanceFlags{sourceRepo: "elastic/\"repo\"\ninjected", sourceFile: "a\r\n.yaml"},
			want:     terraformer.SourceInfo{GitRepo: "elastic/repoinjected", SourceFile: "a.yaml"},
		},
		{
			name:     "stdin with repo override only leaves file empty",
			specPath: "-",
			flags:    provenanceFlags{sourceRepo: "elastic/harp-plugins"},
			want:     terraformer.SourceInfo{GitRepo: "elastic/harp-plugins"},
		},
		{
			name:     "file with both overrides skips git",
			specPath: "specs/a.yaml",
			flags:    provenanceFlags{sourceRepo: "elastic/harp-plugins", sourceFile: "other.yaml"},
			git:      gitOK,
			want:     terraformer.SourceInfo{GitRepo: "elastic/harp-plugins", SourceFile: "other.yaml"},
		},
		{
			name:       "file with repo override takes the file from git",
			specPath:   "specs/a.yaml",
			flags:      provenanceFlags{sourceRepo: "elastic/harp-plugins"},
			git:        gitOK,
			want:       terraformer.SourceInfo{GitRepo: "elastic/harp-plugins", SourceFile: "specs/a.yaml"},
			wantCalled: true,
		},
		{
			name:       "file with file override takes the repo from git",
			specPath:   "specs/a.yaml",
			flags:      provenanceFlags{sourceFile: "other.yaml"},
			git:        gitOK,
			want:       terraformer.SourceInfo{GitRepo: "git/repo", SourceFile: "other.yaml"},
			wantCalled: true,
		},
		{
			name:       "file without overrides uses git",
			specPath:   "specs/a.yaml",
			git:        gitOK,
			want:       terraformer.SourceInfo{GitRepo: "git/repo", SourceFile: "specs/a.yaml"},
			wantCalled: true,
		},
		{
			name:       "git error without the flag warns and keeps overrides",
			specPath:   "specs/a.yaml",
			flags:      provenanceFlags{sourceRepo: "elastic/harp-plugins"},
			git:        gitResult{err: errors.New("not a git repo")},
			want:       terraformer.SourceInfo{GitRepo: "elastic/harp-plugins"},
			wantCalled: true,
		},
		{
			name:       "git error without the flag yields zero value",
			specPath:   "specs/a.yaml",
			git:        gitResult{err: terraformer.ErrGitNotFound},
			want:       terraformer.SourceInfo{},
			wantCalled: true,
		},
		{
			name:     "required with stdin and no overrides fails with the stdin reason",
			specPath: "-",
			flags:    provenanceFlags{required: true},
			want:     terraformer.SourceInfo{},
			wantErr:  errSpecFromStdin,
		},
		{
			name:     "required with stdin and both overrides succeeds",
			specPath: "-",
			flags:    provenanceFlags{required: true, sourceRepo: "elastic/harp-plugins", sourceFile: "a.yaml"},
			want:     terraformer.SourceInfo{GitRepo: "elastic/harp-plugins", SourceFile: "a.yaml"},
		},
		{
			name:     "required with stdin and repo override only is incomplete",
			specPath: "-",
			flags:    provenanceFlags{required: true, sourceRepo: "elastic/harp-plugins"},
			want:     terraformer.SourceInfo{GitRepo: "elastic/harp-plugins"},
			wantErr:  errProvenanceIncomplete,
		},
		{
			name:       "required with git not installed reports git not found",
			specPath:   "specs/a.yaml",
			flags:      provenanceFlags{required: true},
			git:        gitResult{err: terraformer.ErrGitNotFound},
			wantErr:    terraformer.ErrGitNotFound,
			wantCalled: true,
		},
		{
			name:       "required outside a repo reports not a git repo",
			specPath:   "specs/a.yaml",
			flags:      provenanceFlags{required: true},
			git:        gitResult{err: terraformer.ErrNotGitRepo},
			wantErr:    terraformer.ErrNotGitRepo,
			wantCalled: true,
		},
		{
			name:       "required with a repo lacking origin is incomplete",
			specPath:   "specs/a.yaml",
			flags:      provenanceFlags{required: true},
			git:        gitResult{src: terraformer.SourceInfo{SourceFile: "specs/a.yaml"}},
			want:       terraformer.SourceInfo{SourceFile: "specs/a.yaml"},
			wantErr:    errProvenanceIncomplete,
			wantCalled: true,
		},
		{
			name:       "required with git success passes",
			specPath:   "specs/a.yaml",
			flags:      provenanceFlags{required: true},
			git:        gitOK,
			want:       terraformer.SourceInfo{GitRepo: "git/repo", SourceFile: "specs/a.yaml"},
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := resolveGitContext
			t.Cleanup(func() { resolveGitContext = orig })

			called := false
			resolveGitContext = func(context.Context, string) (terraformer.SourceInfo, error) {
				called = true
				return tt.git.src, tt.git.err
			}

			got, err := resolveSourceInfo(context.Background(), tt.specPath, tt.flags)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("resolveSourceInfo() unexpected error = %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("resolveSourceInfo() error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("resolveSourceInfo() = %+v, want %+v", got, tt.want)
			}
			if called != tt.wantCalled {
				t.Errorf("ResolveGitContext called = %v, want %v", called, tt.wantCalled)
			}
		})
	}
}

func Test_addProvenanceFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "x"}
	flags := addProvenanceFlags(cmd)

	args := []string{"--source-repo", "elastic/harp-plugins", "--source-file", "specs/a.yaml", "--require-provenance"}
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}
	want := provenanceFlags{sourceRepo: "elastic/harp-plugins", sourceFile: "specs/a.yaml", required: true}
	if *flags != want {
		t.Errorf("flags = %+v, want %+v", *flags, want)
	}

	fresh := addProvenanceFlags(&cobra.Command{Use: "y"})
	if *fresh != (provenanceFlags{}) {
		t.Errorf("defaults = %+v, want zero value (opt-in)", *fresh)
	}
}

func Test_subcommands_registerProvenanceFlags(t *testing.T) {
	for name, build := range map[string]func() *cobra.Command{
		"service": terraformerServiceCmd,
		"agent":   terraformerAgentCmd,
		"policy":  terraformerPolicyCmd,
	} {
		t.Run(name, func(t *testing.T) {
			c := build()
			for _, f := range []string{"source-repo", "source-file", "require-provenance"} {
				if c.Flags().Lookup(f) == nil {
					t.Errorf("%s: flag --%s not registered", name, f)
				}
			}
		})
	}
}
