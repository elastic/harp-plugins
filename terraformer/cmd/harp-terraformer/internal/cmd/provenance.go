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
	// errInvalidOverride is returned when a --source-repo or --source-file value
	// contains characters that cannot be embedded in the header.
	errInvalidOverride = errors.New("source override contains unsupported characters")
)

// resolveGitContext is a seam so tests can exercise resolveSourceInfo without
// git. Tests that replace it must not run in parallel.
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

// mustResolveSourceInfo is resolveSourceInfo for the subcommands: it exits the
// process when provenance cannot be resolved. It must run before the output is
// opened, so that a failure does not truncate an existing generated file.
func mustResolveSourceInfo(ctx context.Context, specPath string, flags provenanceFlags) terraformer.SourceInfo {
	src, err := resolveSourceInfo(ctx, specPath, flags)
	if err != nil {
		log.For(ctx).Fatal("unable to resolve source provenance", zap.Error(err), zap.String("path", specPath))
	}
	return src
}

// sanitizeOverride returns v sanitized, or errInvalidOverride when sanitizing
// would change it, so a value is never altered silently.
func sanitizeOverride(flag, v string) (string, error) {
	clean := terraformer.SanitizeSourceValue(v)
	if clean != v {
		return "", fmt.Errorf("%w: --%s=%q", errInvalidOverride, flag, v)
	}
	return clean, nil
}

// resolveSourceInfo derives repo and file provenance for specPath. Explicit
// flags win per field; git discovery fills the remaining fields when the
// specification is a real file.
//
// Without --require-provenance a failure to discover provenance is logged and
// the header fields are omitted. With it, the specific reason is returned.
func resolveSourceInfo(ctx context.Context, specPath string, flags provenanceFlags) (terraformer.SourceInfo, error) {
	repo, err := sanitizeOverride("source-repo", flags.sourceRepo)
	if err != nil {
		return terraformer.SourceInfo{}, err
	}
	file, err := sanitizeOverride("source-file", flags.sourceFile)
	if err != nil {
		return terraformer.SourceInfo{}, err
	}
	src := terraformer.SourceInfo{GitRepo: repo, SourceFile: file}

	fromStdin := specPath == "-"
	var discoverErr error
	if !fromStdin && (src.GitRepo == "" || src.SourceFile == "") {
		discoverErr = discoverProvenance(ctx, specPath, &src, flags.required)
	}

	if !flags.required {
		return src, nil
	}
	return src, checkRequired(src, fromStdin, discoverErr)
}

// discoverProvenance fills the empty fields of src from git. A failure is
// returned, and logged unless provenance is required, in which case the caller
// reports it.
func discoverProvenance(ctx context.Context, specPath string, src *terraformer.SourceInfo, required bool) error {
	found, err := resolveGitContext(ctx, specPath)
	if err != nil {
		if !required {
			msg := "source provenance unavailable, header fields omitted"
			if errors.Is(err, terraformer.ErrGitNotFound) {
				msg = "git is not installed, source provenance header fields omitted"
			}
			log.For(ctx).Warn(msg, zap.Error(err))
		}
		return err
	}
	if src.GitRepo == "" {
		src.GitRepo = found.GitRepo
	}
	if src.SourceFile == "" {
		src.SourceFile = found.SourceFile
	}
	return nil
}

// checkRequired reports why src lacks the repo or file, or nil when complete.
func checkRequired(src terraformer.SourceInfo, fromStdin bool, discoverErr error) error {
	var missing []string
	if src.GitRepo == "" {
		missing = append(missing, "repo")
	}
	if src.SourceFile == "" {
		missing = append(missing, "file")
	}
	if len(missing) == 0 {
		return nil
	}

	switch {
	case discoverErr != nil:
		return fmt.Errorf("resolve source provenance: %w", discoverErr)
	case fromStdin && len(missing) == 2:
		return fmt.Errorf("%w (use --source-repo and --source-file)", errSpecFromStdin)
	default:
		return fmt.Errorf("%w: missing %s", errProvenanceIncomplete, strings.Join(missing, " and "))
	}
}
