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
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp-plugins/cmd/harp-assertion/internal/config"
	tasks "github.com/elastic/harp-plugins/cmd/harp-assertion/pkg/tasks/assertion"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
)

// -----------------------------------------------------------------------------

var signCmd = func(conf *config.Configuration) *cobra.Command {
	var (
		inputPath      string
		outputPath     string
		transitKeyName string
		expiration     time.Duration
		audiences      []string
		subject        string
	)

	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Sign a JSON object",
		Run: func(cmd *cobra.Command, args []string) {
			// Prepare command context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-assertion-sign", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Initialize Vault client
			vaultClient, err := api.NewClient(api.DefaultConfig())
			if err != nil {
				log.For(ctx).Fatal("unable to connect to vault", zap.Error(err))
			}

			// Prepare task
			t := tasks.Sign{
				VaultClient:    vaultClient,
				Expiration:     expiration,
				Audiences:      audiences,
				Subject:        subject,
				TransitKeyName: transitKeyName,
				ContentReader:  cmdutil.FileReader(inputPath),
				OutputWriter:   cmdutil.FileWriter(outputPath),
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
	cmd.Flags().StringVar(&transitKeyName, "key", "", "Specify transit key name")
	cmd.Flags().DurationVar(&expiration, "expiration", 5*time.Minute, "Specify assertion expiration time window")
	cmd.Flags().StringVar(&subject, "subject", "", "Specify subject for assertion")
	cmd.Flags().StringArrayVar(&audiences, "audience", []string{}, "Specify target audience for assertion")

	return cmd
}
