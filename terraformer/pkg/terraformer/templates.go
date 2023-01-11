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

import terraformerv1 "github.com/elastic/harp-plugins/terraformer/api/gen/go/harp/terraformer/v1"

// ServiceTemplate is the TF 0.12 Service template.
const ServiceTemplate = `# Generated with Harp Terraformer, Don't modify.
# https://github.com/elastic/harp-plugins/tree/main/cmd/harp-terraformer
# ---
# SpecificationHash: "{{.SpecHash}}"
# Owner: "{{.Meta.Owner}}"
# Date: "{{.Date}}"
# Description: "{{.Meta.Description}}"
# Issues:{{range .Meta.Issues}}
# - {{.}}{{ end }}
# ---
#
# ------------------------------------------------------------------------------

# Create the policy
data "vault_policy_document" "service-{{.ObjectName}}" {
{{- range $ns, $secrets := .Namespaces }}
  # {{ $ns }} secrets{{ range $k, $item := $secrets }}
  rule {
	description  = "{{$item.Description}}"
	path         = "{{$item.Path}}"
    capabilities = [{{range $i, $v := $item.Capabilities}}{{if $i}} ,{{end}}{{printf "%q" $v}}{{end}}]
  }
  {{end -}}
{{end}}{{if .CustomRules }}
  # Custom secret paths{{ range $k, $item := .CustomRules }}
  rule {
	description  = "{{$item.Description}}"
	path         = "{{$item.Path}}"
    capabilities = [{{range $i, $v := $item.Capabilities}}{{if $i}} ,{{end}}{{printf "%q" $v}}{{end}}]
  }
  {{end}}{{end -}}
}

# Register the policy
resource "vault_policy" "service-{{.ObjectName}}" {
  name   = "service-{{.ObjectName}}"
  policy = data.vault_policy_document.service-{{.ObjectName}}.hcl
}

# ------------------------------------------------------------------------------
#
# Register the backend role
resource "vault_approle_auth_backend_role" "{{.ObjectName}}" {
  backend   = "service"
  role_name = "{{.ObjectName}}"

  token_policies = [
	"cso-default",
	"service-default",
    "service-{{.ObjectName}}",
  ]
}
`

// AgentTemplate is the TF 0.12 Agent template.
const AgentTemplate = `# Generated with Harp Terraformer, Don't modify.
# https://github.com/elastic/harp-plugins/tree/main/cmd/harp-terraformer
# ---
# SpecificationHash: "{{.SpecHash}}"
# Owner: "{{.Meta.Owner}}"
# Date: "{{.Date}}"
# Description: "{{.Meta.Description}}"
# Issues:{{range .Meta.Issues}}
# - {{.}}{{ end }}
# ---
#
# ------------------------------------------------------------------------------

# Create the policy
data "vault_policy_document" "agent-{{.ObjectName}}" {

  rule {
    description  = "Allow agent to retrieve service role-id"
	path         = "auth/service/role/{{.ObjectName}}/role-id"
	capabilities = ["read"]
  }

  rule {
	description      = "Allow agent to retrieve secret-id"
	path             = "auth/service/role/{{.ObjectName}}/secret-id"
	capabilities     = ["create", "update"]{{ if not .DisableTokenWrap }}
	min_wrapping_ttl = "1s"  # minimum allowed TTL that clients can specify for a wrapped response
	max_wrapping_ttl = "90s" # maximum allowed TTL that clients can specify for a wrapped response{{end}}
  }
}

# Register the policy
resource "vault_policy" "agent-{{.ObjectName}}" {
  name   = "agent-{{.ObjectName}}"
  policy = data.vault_policy_document.agent-{{.ObjectName}}.hcl
}

# ------------------------------------------------------------------------------
#
# Register the backend role
resource "vault_approle_auth_backend_role" "agent-{{.ObjectName}}" {
  backend   = "agent"
  role_name = "{{.ObjectName}}"

  token_policies = [
	"cso-default",
	"agent-default",
    "agent-{{.ObjectName}}",
  ]
}
`

// -----------------------------------------------------------------------------

type tmplModel struct {
	// SpecHash contains base64 encoded sha256 hash of input specification.
	SpecHash string
	// Meta contains specification metadata
	Meta *terraformerv1.AppRoleDefinitionMeta
	// Date contains the generation data as RFC822 string.
	Date string
	// Environment value
	Environment string
	// Generated role name
	RoleName string
	// Generated object name
	ObjectName string
	// Namespaces and secret paths
	Namespaces map[string][]tmpSecretModel
	// Non CSO bound secret path
	CustomRules []tmpSecretModel
	// DisableTokenWrap disable token wrap enforcement
	DisableTokenWrap bool
}

type tmpSecretModel struct {
	Path         string
	Description  string
	Capabilities []string
}
