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
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	iconfig "github.com/elastic/harp-plugins/terraformer/cmd/harp-terraformer/internal/config"
	"github.com/elastic/harp/build/version"
	"github.com/elastic/harp/pkg/sdk/config"
	configcmd "github.com/elastic/harp/pkg/sdk/config/cmd"
	"github.com/elastic/harp/pkg/sdk/log"
)

// -----------------------------------------------------------------------------

var (
	cfgFile string
	conf    = &iconfig.Configuration{}
)

const (
	envPrefix = "SHUB_TERRAFORMER"
)

// -----------------------------------------------------------------------------

// RootCmd describes root command of the tool
var mainCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "harp-terraformer",
		Short: "Harp CSO Vault policy generator",
	}

	// Register falgs
	cmd.Flags().StringVar(&cfgFile, "config", "", "config file")

	// Register sub commands
	cmd.AddCommand(version.Command())
	cmd.AddCommand(configcmd.NewConfigCommand(conf, envPrefix))

	// Add subcommands
	cmd.AddCommand(terraformerAgentCmd())
	cmd.AddCommand(terraformerApproleCmd())
	cmd.AddCommand(terraformerPolicyCmd())
	cmd.AddCommand(terraformerServiceCmd())

	// Return command
	return cmd
}

// -----------------------------------------------------------------------------

// Execute main command
func Execute() error {
	initConfig()

	// Initialize root command
	cmd := mainCmd()
	return cmd.Execute()
}

// -----------------------------------------------------------------------------

func initConfig() {
	if err := config.Load(conf, envPrefix, cfgFile); err != nil {
		log.Bg().Fatal("Unable to load settings", zap.Error(err))
	}
}
