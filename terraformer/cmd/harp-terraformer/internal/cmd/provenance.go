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
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp-plugins/terraformer/pkg/terraformer"
	"github.com/elastic/harp/pkg/sdk/log"
)

var (
	// errSpecFromStdin is returned when provenance is required but the
	// specification comes from stdin and no override was given.
	errSpecFromStdin = errors.New("specification is read from stdin, so repo and file cannot be derived")
	// errProvenanceIncomplete is returned when provenance is required but the
	// repo or file could not be determined.
	errProvenanceIncomplete = errors.New("source provenance is incomplete")
)

// resolveGitContext is a seam so tests can exercise resolveSourceInfo without git.
var resolveGitContext = terraformer.ResolveGitContext

// provenanceFlags holds the provenance options shared by every subcommand.
type provenanceFlags struct {
	sourceRepo string
	sourceFile string
	required   bool
}

// addProvenanceFlags registers the provenance flags on cmd, so all
// subcommands expose the same options.
func addProvenanceFlags(cmd *cobra.Command) *provenanceFlags {
	f := &provenanceFlags{}
	cmd.Flags().StringVar(&f.sourceRepo, "source-repo", "", "Repository name recorded in the header, overriding git discovery (e.g. 'elastic/harp-plugins')")
	cmd.Flags().StringVar(&f.sourceFile, "source-file", "", "Repo-relative specification path recorded in the header, overriding git discovery")
	cmd.Flags().BoolVar(&f.required, "require-provenance", false, "Fail when the repo and file cannot be determined")
	return f
}

// resolveSourceInfo derives repo and file provenance for specPath. Explicit
// flags win per field; git discovery fills the remaining fields when the
// specification is a real file.
//
// Without --require-provenance a failure to discover provenance is logged and
// the header fields are omitted. With it, the specific reason is returned.
func resolveSourceInfo(ctx context.Context, specPath string, flags provenanceFlags) (terraformer.SourceInfo, error) {
	src := terraformer.SourceInfo{
		GitRepo:    terraformer.SanitizeSourceValue(flags.sourceRepo),
		SourceFile: terraformer.SanitizeSourceValue(flags.sourceFile),
	}
	fromStdin := specPath == "-"

	var discoverErr error
	if !fromStdin && (src.GitRepo == "" || src.SourceFile == "") {
		found, err := resolveGitContext(ctx, specPath)
		switch {
		case err != nil:
			discoverErr = err
			if !flags.required {
				msg := "source provenance unavailable, header fields omitted"
				if errors.Is(err, terraformer.ErrGitNotFound) {
					msg = "git is not installed, source provenance header fields omitted"
				}
				log.For(ctx).Warn(msg, zap.Error(err))
			}
		default:
			if src.GitRepo == "" {
				src.GitRepo = found.GitRepo
			}
			if src.SourceFile == "" {
				src.SourceFile = found.SourceFile
			}
		}
	}

	if !flags.required {
		return src, nil
	}

	var missing []string
	if src.GitRepo == "" {
		missing = append(missing, "repo")
	}
	if src.SourceFile == "" {
		missing = append(missing, "file")
	}
	if len(missing) == 0 {
		return src, nil
	}

	switch {
	case discoverErr != nil:
		return src, discoverErr
	case fromStdin && len(missing) == 2:
		return src, fmt.Errorf("%w (use --source-repo and --source-file)", errSpecFromStdin)
	default:
		return src, fmt.Errorf("%w: missing %s", errProvenanceIncomplete, strings.Join(missing, " and "))
	}
}
