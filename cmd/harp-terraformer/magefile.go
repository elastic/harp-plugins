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

//+build mage

package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/elastic/harp/build/artifact"
	"github.com/elastic/harp/build/mage/git"
	"github.com/elastic/harp/build/mage/golang"
	"github.com/elastic/harp/build/mage/release"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	"github.com/magefile/mage/mg"
)

var Default = Build

var descriptor = &artifact.Command{
	Package:     "github.com/elastic/harp-plugins",
	Module:      "cmd/harp-terraformer",
	Name:        "Harp Terraformer",
	Description: "Harp CSO Vault Policy generator",
}

// Build the artefact
func Build() {
	banner := figure.NewFigure("Harp Terraformer", "", true)
	banner.Print()

	fmt.Println("")
	color.Red("# Build Info ---------------------------------------------------------------")
	fmt.Printf("Go version : %s\n", runtime.Version())

	color.Red("# Pipeline -----------------------------------------------------------------")
	mg.SerialDeps(golang.Vendor, golang.License("../../"), golang.Lint("../../"), Test)

	color.Red("# Artefact(s) --------------------------------------------------------------")
	mg.Deps(Compile)
}

// Test application
func Test() {
	color.Cyan("## Tests")
	mg.SerialDeps(
		func() error {
			return golang.UnitTest("github.com/elastic/harp-plugins/cmd/harp-terraformer/internal/...")()
		},
		func() error {
			return golang.UnitTest("github.com/elastic/harp-plugins/cmd/harp-terraformer/pkg/...")()
		},
	)
}

// Compile artefacts
func Compile() error {
	// Extract
	version, err := git.TagMatch("cmd/harp-terraformer/v*")
	if err != nil {
		return err
	}

	// Build artifact
	return golang.Build("harp-terraformer", "github.com/elastic/harp-plugins/cmd/harp-terraformer", version)()
}

// Release
func Release(ctx context.Context) error {
	color.Red(fmt.Sprintf("# Releasing (%s) ---------------------------------------------------------------", os.Getenv("RELEASE")))

	color.Cyan("## Cross compiling artifact")

	// Extract
	version, err := git.TagMatch("cmd/harp-terraformer/v*")
	if err != nil {
		return err
	}

	mg.CtxDeps(ctx,
		func() error {
			return golang.Release(
				"harp-terraformer",
				"github.com/elastic/harp-plugins/cmd/harp-terraformer",
				version,
				golang.GOOS("darwin"), golang.GOARCH("amd64"),
			)()
		},
		func() error {
			return golang.Release(
				"harp-terraformer",
				"github.com/elastic/harp-plugins/cmd/harp-terraformer",
				version,
				golang.GOOS("linux"), golang.GOARCH("amd64"),
			)()
		},
		func() error {
			return golang.Release(
				"harp-terraformer",
				"github.com/elastic/harp-plugins/cmd/harp-terraformer",
				version,
				golang.GOOS("windows"), golang.GOARCH("amd64"),
			)()
		},
	)

	// No error
	return ctx.Err()
}

func Homebrew() error {
	return release.HomebrewFormula(descriptor)()
}
