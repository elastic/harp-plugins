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

//go:build mage
// +build mage

package main

import (
	"github.com/magefile/mage/mg"

	"github.com/elastic/harp/build/artifact"
	"github.com/elastic/harp/build/mage/docker"
)

var (
	harpServer = &artifact.Command{
		Package:     "github.com/elastic/harp-plugins",
		Module:      "cmd/harp-server",
		Name:        "Harp Server",
		Description: "Harp Container Server",
	}
	harpTerraformer = &artifact.Command{
		Package:     "github.com/elastic/harp-plugins",
		Module:      "cmd/harp-terraformer",
		Name:        "Harp Terraformer",
		Description: "Harp CSO Vault Policy generator",
	}
)

// -----------------------------------------------------------------------------

type Releaser mg.Namespace

// Server builds the harp-server binaries using docker pipeline.
func (Releaser) Server() error {
	return docker.Release(harpServer)()
}

// Terraformer builds the harp-terraformer binaries using docker pipeline.
func (Releaser) Terraformer() error {
	return docker.Release(harpTerraformer)()
}
