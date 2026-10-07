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

	"github.com/elastic/harp-plugins/terraformer/pkg/terraformer"
)

func Test_resolveSourceInfo(t *testing.T) {
	want := terraformer.SourceInfo{GitRepo: "elastic/harp-plugins", SourceFile: "specs/a.yaml"}

	tests := []struct {
		name       string
		specPath   string
		resolve    func(context.Context, string) (terraformer.SourceInfo, error)
		want       terraformer.SourceInfo
		wantCalled bool
	}{
		{
			name:     "stdin skips git",
			specPath: "-",
			resolve: func(context.Context, string) (terraformer.SourceInfo, error) {
				return want, nil
			},
			want:       terraformer.SourceInfo{},
			wantCalled: false,
		},
		{
			name:     "resolution error yields zero value",
			specPath: "specs/a.yaml",
			resolve: func(context.Context, string) (terraformer.SourceInfo, error) {
				return terraformer.SourceInfo{}, errors.New("not a git repo")
			},
			want:       terraformer.SourceInfo{},
			wantCalled: true,
		},
		{
			name:     "git not installed yields zero value",
			specPath: "specs/a.yaml",
			resolve: func(context.Context, string) (terraformer.SourceInfo, error) {
				return terraformer.SourceInfo{}, terraformer.ErrGitNotFound
			},
			want:       terraformer.SourceInfo{},
			wantCalled: true,
		},
		{
			name:     "success is passed through",
			specPath: "specs/a.yaml",
			resolve: func(context.Context, string) (terraformer.SourceInfo, error) {
				return want, nil
			},
			want:       want,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := resolveGitContext
			t.Cleanup(func() { resolveGitContext = orig })

			called := false
			resolveGitContext = func(ctx context.Context, p string) (terraformer.SourceInfo, error) {
				called = true
				return tt.resolve(ctx, p)
			}

			got := resolveSourceInfo(context.Background(), tt.specPath)
			if got != tt.want {
				t.Errorf("resolveSourceInfo() = %+v, want %+v", got, tt.want)
			}
			if called != tt.wantCalled {
				t.Errorf("ResolveGitContext called = %v, want %v", called, tt.wantCalled)
			}
		})
	}
}
