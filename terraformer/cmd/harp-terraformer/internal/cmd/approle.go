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
	"io"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp-plugins/terraformer/pkg/terraformer"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
)

var (
	terraformerApproleInputSpec                string
	terraformerApproleOutputPath               string
	terraformerApproleDisableTokenWrap         bool
	terraformerApproleDisableEnvironmentSuffix bool
	terraformerApproleEnvironment              string
)

// -----------------------------------------------------------------------------

var terraformerApproleCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approle",
		Short: "policy and approle with approle engine",
		Run:   runTerraformerApprole,
	}

	// Parameters
	cmd.Flags().StringVar(&terraformerApproleInputSpec, "spec", "-", "AppRole specification path ('-' for stdin or filename)")
	cmd.Flags().StringVar(&terraformerApproleOutputPath, "out", "-", "Output file ('-' for stdout or a filename)")
	cmd.Flags().StringVar(&terraformerApproleEnvironment, "env", "production", "Target environment")
	cmd.Flags().BoolVar(&terraformerApproleDisableTokenWrap, "no-token-wrap", false, "Disable token wrapping")
	cmd.Flags().BoolVar(&terraformerApproleDisableEnvironmentSuffix, "no-env-suffix", false, "Disable environment suffix in role and policy names")

	return cmd
}

func runTerraformerApprole(cmd *cobra.Command, _ []string) {
	ctx, cancel := cmdutil.Context(cmd.Context(), "harp-terraformer-approle", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
	defer cancel()

	var (
		reader io.Reader
		err    error
	)

	// Create input reader
	reader, err = cmdutil.Reader(terraformerApproleInputSpec)
	if err != nil {
		log.For(ctx).Fatal("unable to open input specification", zap.Error(err), zap.String("path", terraformerApproleInputSpec))
	}

	// Create output writer
	writer, err := cmdutil.Writer(terraformerApproleOutputPath)
	if err != nil {
		log.For(ctx).Fatal("unable to create output writer", zap.Error(err), zap.String("path", terraformerApproleOutputPath))
	}

	// Run terraformer
	if err := terraformer.Run(ctx, reader, terraformerApproleEnvironment, terraformerApproleDisableTokenWrap, terraformerApproleDisableEnvironmentSuffix, terraformer.ApproleTemplate, writer); err != nil {
		log.For(ctx).Fatal("unable to process specification", zap.Error(err), zap.String("path", terraformerApproleInputSpec))
	}
}
