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

	"go.uber.org/zap"

	"github.com/elastic/harp-plugins/terraformer/pkg/terraformer"
	"github.com/elastic/harp/pkg/sdk/log"
)

// resolveGitContext is a seam so tests can exercise resolveSourceInfo without git.
var resolveGitContext = terraformer.ResolveGitContext

// resolveSourceInfo attempts to derive repo and file provenance for specPath.
// Returns a zero-value SourceInfo when specPath is stdin ("-") or git is unavailable.
func resolveSourceInfo(ctx context.Context, specPath string) terraformer.SourceInfo {
	if specPath == "-" {
		return terraformer.SourceInfo{}
	}
	src, err := resolveGitContext(ctx, specPath)
	if err != nil {
		msg := "source provenance unavailable, header fields omitted"
		if errors.Is(err, terraformer.ErrGitNotFound) {
			msg = "git is not installed, source provenance header fields omitted"
		}
		log.For(ctx).Warn(msg, zap.Error(err))
		return terraformer.SourceInfo{}
	}
	return src
}
