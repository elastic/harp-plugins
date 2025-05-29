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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v2"

	terraformerv1 "github.com/elastic/harp-plugins/terraformer/api/gen/go/harp/terraformer/v1"
	csov1 "github.com/elastic/harp/pkg/cso/v1"
	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/template/engine"
	vpath "github.com/elastic/harp/pkg/vault/path"
)

type suffixProcessorMap map[csov1.Ring]struct {
	prefix     []string
	suffixFunc func() []*terraformerv1.AppRoleDefinitionSecretSuffix
}

func buildNamespaces(def *terraformerv1.AppRoleDefinition, env string) (suffixProcessorMap, error) {
	// Check arguments
	if err := validate(def); err != nil {
		return nil, err
	}

	// Build processors map
	return map[csov1.Ring]struct {
		prefix     []string
		suffixFunc func() []*terraformerv1.AppRoleDefinitionSecretSuffix
	}{
		csov1.RingInfra: {
			suffixFunc: def.Spec.Namespaces.GetInfrastructure,
		},
		csov1.RingPlatform: {
			prefix:     []string{env, def.Spec.Selector.Platform},
			suffixFunc: def.Spec.Namespaces.GetPlatform,
		},
		csov1.RingProduct: {
			prefix:     []string{def.Spec.Selector.Product, def.Spec.Selector.Version, def.Spec.Selector.Component},
			suffixFunc: def.Spec.Namespaces.GetProduct,
		},
		csov1.RingApplication: {
			prefix:     []string{env, def.Spec.Selector.Platform, def.Spec.Selector.Product, def.Spec.Selector.Version, def.Spec.Selector.Component},
			suffixFunc: def.Spec.Namespaces.GetApplication,
		},
		csov1.RingArtifact: {
			suffixFunc: def.Spec.Namespaces.GetArtifact,
		},
	}, nil
}

// Config to CSO transformer
func pathCompiler(ring csov1.Ring, prefix []string, suffixFunc func() []*terraformerv1.AppRoleDefinitionSecretSuffix, res *tmplModel) error {
	// Retrieve suffix list
	secretSuffixList := suffixFunc()

	// Check nil / len
	if len(secretSuffixList) == 0 {
		return nil
	}

	// Foreach suffix
	for _, item := range secretSuffixList {
		// Check arguments
		if item == nil {
			continue
		}

		// Convert definition to CSO secret path
		v, err := ring.Path(append(prefix, item.Suffix)...)
		if err != nil {
			// Will be used after path validation control implementation
			return fmt.Errorf("unable to extract ring from path: %w", err)
		}

		// Check description
		if item.Description == "" {
			return fmt.Errorf("missing description for secret suffix '%s'", v)
		}

		// Filter capabilities
		capabilities := types.StringArray(filterCapabilities(item.Capabilities))

		// Add metadata access for list operation
		if capabilities.Contains("list") {
			// Add to mapped secrets
			res.Namespaces[ring.Name()] = append(res.Namespaces[ring.Name()], tmpSecretModel{
				Path:         vaultKvV2Path(v, "metadata"),
				Description:  "Allow metadata access for list operation",
				Capabilities: []string{"list"},
			})

			// Remove "list" from capabilities
			capabilities.Remove("list")
		}

		// Add to mapped secrets
		res.Namespaces[ring.Name()] = append(res.Namespaces[ring.Name()], tmpSecretModel{
			Path:         vaultKvV2Path(v, "data"),
			Description:  item.Description,
			Capabilities: capabilities,
		})
	}

	// No error
	return nil
}

func compile(env string, def *terraformerv1.AppRoleDefinition, specHash string, noTokenWrap bool) (*tmplModel, error) {
	// Check arguments
	if err := validate(def); err != nil {
		return nil, err
	}

	res := &tmplModel{
		Date:             time.Now().UTC().Format(time.RFC3339),
		SpecHash:         specHash,
		Meta:             def.Meta,
		Environment:      slug.Make(env),
		RoleName:         slug.Make(def.Meta.Name),
		ObjectName:       slug.Make(fmt.Sprintf("%s %s", def.Meta.Name, env)),
		Namespaces:       map[string][]tmpSecretModel{},
		DisableTokenWrap: noTokenWrap,
	}

	if def.Spec.Namespaces != nil {
		// Prepare ring processors
		csoNamespaces, err := buildNamespaces(def, env)
		if err != nil {
			return nil, err
		}

		// Process CSO rings
		for ns, suffixes := range csoNamespaces {
			if err := pathCompiler(ns, suffixes.prefix, suffixes.suffixFunc, res); err != nil {
				return nil, err
			}
		}
	}

	// Process custom paths
	customRules := []tmpSecretModel{}
	for _, customRule := range def.Spec.Custom {
		// Check arguments
		if customRule == nil {
			continue
		}

		// Check suffix
		if customRule.Suffix == "" {
			return nil, fmt.Errorf("missing suffix for secret")
		}

		// Check description
		if customRule.Description == "" {
			return nil, fmt.Errorf("missing description for secret suffix '%s'", customRule.Suffix)
		}

		customPath, err := engine.Render(customRule.Suffix, map[string]interface{}{
			"Env":      res.Environment,
			"Selector": def.Spec.Selector,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to compile suffix template: %w", err)
		}

		customRules = append(customRules, tmpSecretModel{
			Path:         vpath.SanitizePath(customPath),
			Description:  customRule.Description,
			Capabilities: filterCapabilities(customRule.Capabilities),
		})
	}

	// Assign result
	if len(customRules) > 0 {
		res.CustomRules = customRules
	}

	return res, nil
}

// List of allowed capabilities
var allowedCapabilities = types.StringArray{
	"list",
	"create",
	"read",
	"update",
	"delete",
	"sudo",
}

// filterCapabilities removes useless capabilities
func filterCapabilities(list []string) []string {
	// Default to read
	res := types.StringArray([]string{"read"})

	// Iterate of given capabilities
	for _, c := range list {
		// If capability is not allowed
		if !allowedCapabilities.Contains(c) {
			continue
		}
		// Add to result
		res.AddIfNotContains(c)
	}

	// Return filtered capabilities
	return res
}

// loadFromYAML reads YAML definition and returns the PB struct.
//
// Protobuf doesn't contain YAML struct tags and json one are not symetric
// to protobuf. We need to export YAML as JSON, and then read JSON to Protobuf
// as done in k8s yaml loader.
func loadFromYAML(in []byte) (*terraformerv1.AppRoleDefinition, error) {
	// Decode as YAML any object
	var specBody interface{}
	if err := yaml.Unmarshal(in, &specBody); err != nil {
		return nil, fmt.Errorf("unable to decode YAML input: %w", err)
	}

	// Convert map[interface{}]interface{} to a JSON serializable struct
	specBody, err := convertMapStringInterface(specBody)
	if err != nil {
		return nil, err
	}

	// Marshal as json
	jsonData, err := json.Marshal(specBody)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal to JSON: %w", err)
	}

	// Initialize empty definition object
	def := terraformerv1.AppRoleDefinition{}
	def.Reset()

	// Deserialize JSON with JSONPB wrapper
	if err := protojson.Unmarshal(jsonData, &def); err != nil {
		return nil, fmt.Errorf("unable to decode with ProtoJSON: %w", err)
	}

	// No error
	return &def, nil
}

// Converts map[interface{}]interface{} into map[string]interface{} for json.Marshaler
func convertMapStringInterface(val interface{}) (interface{}, error) {
	switch items := val.(type) {
	case map[interface{}]interface{}:
		result := map[string]interface{}{}
		for k, v := range items {
			key, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("typeError: value %s (type `%s') can't be assigned to type 'string'", k, reflect.TypeOf(k))
			}
			value, err := convertMapStringInterface(v)
			if err != nil {
				return nil, err
			}
			result[key] = value
		}
		return result, nil
	case []interface{}:
		for k, v := range items {
			value, err := convertMapStringInterface(v)
			if err != nil {
				return nil, err
			}
			items[k] = value
		}
	}
	return val, nil
}

func vaultKvV2Path(in, prefix string) string {
	// Add data for kvV2 Vault path
	secretPathParts := strings.SplitN(in, "/", 2)
	return fmt.Sprintf("%s/%s/%s", secretPathParts[0], prefix, secretPathParts[1])
}
