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

package terraformer

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// minimalSpec is a valid AppRoleDefinition with no namespaces.
const minimalSpec = `apiVersion: harp.elastic.co/terraformer/v1
kind: AppRoleDefinition
meta:
  name: "smoke-test"
  owner: "security@elastic.co"
  description: "Minimal spec for tests"
spec:
  selector:
    platform: "security"
    product: "harp"
    version: "v1"
    component: "test"
`

// applicationSpec is a valid AppRoleDefinition with an application namespace.
const applicationSpec = `apiVersion: harp.elastic.co/terraformer/v1
kind: AppRoleDefinition
meta:
  name: "full-test"
  owner: "security@elastic.co"
  description: "Full spec for tests"
spec:
  selector:
    platform: "security"
    product: "harp"
    version: "v1"
    component: "test"
  namespaces:
    application:
      - suffix: "config/db"
        description: "Database config"
        capabilities: ["read"]
`

func Test_Run_sourceInfo_renderedInOutput(t *testing.T) {
	src := SourceInfo{
		GitRepo:    "elastic/harp-plugins",
		SourceFile: "terraformer/spec.yaml",
		GitCommit:  "abc123def456+dirty",
	}

	var out bytes.Buffer
	err := Run(context.Background(), strings.NewReader(minimalSpec), "staging", true, "service", ServiceTemplate, src, &out)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := out.String()
	for _, want := range []string{
		`# GitRepo: "elastic/harp-plugins"`,
		`# SourceFile: "terraformer/spec.yaml"`,
		`# GitCommit: "abc123def456+dirty"`,
	} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, output)
		}
	}
}

func Test_Run_emptySourceInfo_omitsGitLines(t *testing.T) {
	var out bytes.Buffer
	err := Run(context.Background(), strings.NewReader(minimalSpec), "staging", true, "service", ServiceTemplate, SourceInfo{}, &out)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := out.String()
	for _, absent := range []string{"GitRepo", "SourceFile", "GitCommit"} {
		if strings.Contains(output, "# "+absent+":") {
			t.Errorf("output should not contain %q line when SourceInfo is empty\nfull output:\n%s", absent, output)
		}
	}
}

func Test_Run_withApplicationNamespace(t *testing.T) {
	var out bytes.Buffer
	err := Run(context.Background(), strings.NewReader(applicationSpec), "production", true, "service", ServiceTemplate, SourceInfo{}, &out)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "vault_policy_document") {
		t.Errorf("expected vault_policy_document in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Database config") {
		t.Errorf("expected secret description in output, got:\n%s", output)
	}
}

func Test_Run_allTemplates_sourceInfo(t *testing.T) {
	src := SourceInfo{
		GitRepo:    "elastic/harp-plugins",
		SourceFile: "spec.yaml",
		GitCommit:  "deadbeef",
	}

	templates := []struct {
		name     string
		template string
	}{
		{"service", ServiceTemplate},
		{"agent", AgentTemplate},
		{"policy", PolicyTemplate},
	}

	for _, tmpl := range templates {
		t.Run(tmpl.name, func(t *testing.T) {
			var out bytes.Buffer
			err := Run(context.Background(), strings.NewReader(minimalSpec), "staging", true, tmpl.name, tmpl.template, src, &out)
			if err != nil {
				t.Fatalf("Run(%s) error = %v", tmpl.name, err)
			}

			output := out.String()
			if !strings.Contains(output, `# GitRepo: "elastic/harp-plugins"`) {
				t.Errorf("%s template: output missing GitRepo line\nfull output:\n%s", tmpl.name, output)
			}
			if !strings.Contains(output, `# SourceFile: "spec.yaml"`) {
				t.Errorf("%s template: output missing SourceFile line\nfull output:\n%s", tmpl.name, output)
			}
			if !strings.Contains(output, `# GitCommit: "deadbeef"`) {
				t.Errorf("%s template: output missing GitCommit line\nfull output:\n%s", tmpl.name, output)
			}
		})
	}
}

func Test_Run_invalidYAML_returnsError(t *testing.T) {
	var out bytes.Buffer
	err := Run(context.Background(), strings.NewReader("not: valid: yaml: spec"), "staging", true, "service", ServiceTemplate, SourceInfo{}, &out)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}
