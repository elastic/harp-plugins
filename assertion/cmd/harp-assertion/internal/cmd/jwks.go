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
	"crypto/x509"
	"encoding/json"
	"encoding/pem"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp-plugins/assertion/cmd/harp-assertion/internal/config"
	"github.com/elastic/harp-plugins/assertion/pkg/jwtvault"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
)

// -----------------------------------------------------------------------------

var jwksCmd = func(conf *config.Configuration) *cobra.Command {
	var (
		outputPath     string
		transitKeyName string
		pemOutput      bool
	)

	cmd := &cobra.Command{
		Use:   "jwks",
		Short: "Export public keys as JWKS",
		Run: func(cmd *cobra.Command, args []string) {
			// Prepare command context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-assertion-jwks", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Initialize Vault client
			vaultClient, err := api.NewClient(api.DefaultConfig())
			if err != nil {
				log.For(ctx).Fatal("unable to connect to vault", zap.Error(err))
			}

			// Retrieve JWKS from Vault
			vaultJWKS, err := jwtvault.JWKS(vaultClient, "assertions", transitKeyName)
			if err != nil {
				log.For(ctx).Fatal("unable to build JWKS from Vault", zap.Error(err))
			}

			// Get writer
			writer, err := cmdutil.FileWriter(outputPath)(ctx)
			if err != nil {
				log.For(ctx).Fatal("unable to allocate writer", zap.Error(err))
			}

			if pemOutput {
				for idx := range vaultJWKS.Keys {
					// Retrieve key
					k := vaultJWKS.Keys[idx]

					// Marshal to x509
					data, errMarshal := x509.MarshalPKIXPublicKey(k.Key)
					if errMarshal != nil {
						log.For(ctx).Fatal("unable to convert to PKIX", zap.Error(errMarshal))
					}

					// Encode as PEM
					if err = pem.Encode(writer, &pem.Block{
						Type:  "PUBLIC KEY",
						Bytes: data,
					}); err != nil {
						log.For(ctx).Fatal("unable to encode PEM data", zap.Error(err))
					}
				}
			} else {
				// Encode as json
				if err = json.NewEncoder(writer).Encode(vaultJWKS); err != nil {
					log.For(ctx).Fatal("unable to encode JWKS", zap.Error(err))
				}
			}
		},
	}

	// Parameters
	cmd.Flags().StringVar(&outputPath, "out", "-", "Output file ('-' for stdout or a filename)")
	cmd.Flags().StringVar(&transitKeyName, "key", "", "Specify transit key name")
	cmd.Flags().BoolVar(&pemOutput, "pem", false, "Export as PEM")

	return cmd
}
