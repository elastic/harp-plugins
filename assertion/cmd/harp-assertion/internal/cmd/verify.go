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
	"encoding/json"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/square/go-jose.v2"

	"github.com/elastic/harp-plugins/assertion/cmd/harp-assertion/internal/config"
	tasks "github.com/elastic/harp-plugins/assertion/pkg/tasks/assertion"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
)

// -----------------------------------------------------------------------------

var verifyCmd = func(conf *config.Configuration) *cobra.Command {
	var (
		inputPath  string
		outputPath string
		jwksPath   string
	)

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify an assertion and dump content",
		Run: func(cmd *cobra.Command, args []string) {
			// Prepare command context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-assertion-verify", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Get writer
			reader, err := cmdutil.FileReader(jwksPath)(ctx)
			if err != nil {
				log.For(ctx).Fatal("unable to allocate JWKS reader", zap.Error(err))
			}

			// Decode JWKS
			var jwks jose.JSONWebKeySet
			if err := json.NewDecoder(reader).Decode(&jwks); err != nil {
				log.For(ctx).Fatal("unable to decode JWKS", zap.Error(err))
			}

			// Prepare task
			t := tasks.Verify{
				JWKS:          &jwks,
				ContentReader: cmdutil.FileReader(inputPath),
				OutputWriter:  cmdutil.FileWriter(outputPath),
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&inputPath, "in", "-", "Content to sign ('-' for stdin or filename)")
	cmd.Flags().StringVar(&outputPath, "out", "-", "Output file ('-' for stdout or a filename)")
	cmd.Flags().StringVar(&jwksPath, "jwks", "", "JWKS file ('-' for stdout or a filename)")
	log.CheckErr("unable to set 'jwks' flag as required.", cmd.MarkFlagRequired("jwks"))

	return cmd
}
