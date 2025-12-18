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
	"fmt"
	"reflect"
	"testing"

	"github.com/gosimple/slug"

	terraformerv1 "github.com/elastic/harp-plugins/terraformer/api/gen/go/harp/terraformer/v1"
	fuzz "github.com/google/gofuzz"
)

func Test_compile(t *testing.T) {
	type args struct {
		env                   string
		def                   *terraformerv1.AppRoleDefinition
		specHash              string
		noTokenWrap           bool
		defaultAuthEngineName string
	}
	tests := []struct {
		name    string
		args    args
		want    *tmplModel
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "nil meta",
			args: args{
				env: "production",
				def: &terraformerv1.AppRoleDefinition{
					ApiVersion: "harp.elastic.co/terraformer/v1",
					Kind:       "AppRoleDefinition",
				},
				noTokenWrap: false,
				specHash:    "123456",
			},

			wantErr: true,
		},
		{
			name: "missing name in meta",
			args: args{
				env: "production",
				def: &terraformerv1.AppRoleDefinition{
					ApiVersion: "harp.elastic.co/terraformer/v1",
					Kind:       "AppRoleDefinition",
					Meta:       &terraformerv1.AppRoleDefinitionMeta{},
				},
			},
			wantErr: true,
		},
		{
			name: "missing owner in meta",
			args: args{
				env: "production",
				def: &terraformerv1.AppRoleDefinition{
					ApiVersion: "harp.elastic.co/terraformer/v1",
					Kind:       "AppRoleDefinition",
					Meta: &terraformerv1.AppRoleDefinitionMeta{
						Name: "foo",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing description in meta",
			args: args{
				env: "production",
				def: &terraformerv1.AppRoleDefinition{
					ApiVersion: "harp.elastic.co/terraformer/v1",
					Kind:       "AppRoleDefinition",
					Meta: &terraformerv1.AppRoleDefinitionMeta{
						Name:  "foo",
						Owner: "security@elastic.co",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "nil spec",
			args: args{
				env: "production",
				def: &terraformerv1.AppRoleDefinition{
					ApiVersion: "harp.elastic.co/terraformer/v1",
					Kind:       "AppRoleDefinition",
					Meta: &terraformerv1.AppRoleDefinitionMeta{
						Name:        "foo",
						Owner:       "security@elastic.co",
						Description: "test",
					},
				},
				noTokenWrap: false,
				specHash:    "123456",
			},
			wantErr: true,
		},
		{
			name: "empty spec",
			args: args{
				env: "production",
				def: &terraformerv1.AppRoleDefinition{
					ApiVersion: "harp.elastic.co/terraformer/v1",
					Kind:       "AppRoleDefinition",
					Meta: &terraformerv1.AppRoleDefinitionMeta{
						Name:        "foo",
						Owner:       "security@elastic.co",
						Description: "test",
					},
					Spec: &terraformerv1.AppRoleDefinitionSpec{},
				},
				noTokenWrap: false,
				specHash:    "123456",
			},
			wantErr: true,
		},
		{
			name: "empty spec selector",
			args: args{
				env: "production",
				def: &terraformerv1.AppRoleDefinition{
					ApiVersion: "harp.elastic.co/terraformer/v1",
					Kind:       "AppRoleDefinition",
					Meta: &terraformerv1.AppRoleDefinitionMeta{
						Name:        "foo",
						Owner:       "security@elastic.co",
						Description: "test",
					},
					Spec: &terraformerv1.AppRoleDefinitionSpec{
						Selector: &terraformerv1.AppRoleDefinitionSelector{},
					},
				},
				noTokenWrap: false,
				specHash:    "123456",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := compile(tt.args.env, tt.args.def, tt.args.specHash, tt.args.noTokenWrap, tt.args.defaultAuthEngineName)
			if (err != nil) != tt.wantErr {
				t.Errorf("compile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_compile_template_object(t *testing.T) {
	type args struct {
		env                   string
		def                   *terraformerv1.AppRoleDefinition
		specHash              string
		noTokenWrap           bool
		defaultAuthEngineName string
	}
	tests := []struct {
		name    string
		args    args
		want    *tmplModel
		wantErr bool
	}{
		{
			name: "spec disableEnvironmentSuffix=false (default)",
			args: args{
				env: "production",
				def: &terraformerv1.AppRoleDefinition{
					ApiVersion: "harp.elastic.co/terraformer/v1",
					Kind:       "AppRoleDefinition",
					Meta: &terraformerv1.AppRoleDefinitionMeta{
						Name:        "foo",
						Owner:       "security@elastic.co",
						Description: "test",
					},
					Spec: &terraformerv1.AppRoleDefinitionSpec{
						Selector:                 &terraformerv1.AppRoleDefinitionSelector{},
						DisableEnvironmentSuffix: false,
					},
				},
				noTokenWrap: false,
				specHash:    "123456",
			},
			want: &tmplModel{
				Meta: &terraformerv1.AppRoleDefinitionMeta{
					Name:        "foo",
					Owner:       "security@elastic.co",
					Description: "test",
				},
				Environment:              "production",
				RoleName:                 "foo",
				ObjectName:               "foo-production",
				DisableEnvironmentSuffix: false,
			},
			wantErr: false,
		},
		{
			name: "spec disableEnvironmentSuffix=true",
			args: args{
				env: "production",
				def: &terraformerv1.AppRoleDefinition{
					ApiVersion: "harp.elastic.co/terraformer/v1",
					Kind:       "AppRoleDefinition",
					Meta: &terraformerv1.AppRoleDefinitionMeta{
						Name:        "foo",
						Owner:       "security@elastic.co",
						Description: "test",
					},
					Spec: &terraformerv1.AppRoleDefinitionSpec{
						Selector:                 &terraformerv1.AppRoleDefinitionSelector{},
						DisableEnvironmentSuffix: true,
					},
				},
				noTokenWrap: false,
				specHash:    "123456",
			},
			want: &tmplModel{
				Meta: &terraformerv1.AppRoleDefinitionMeta{
					Name:        "foo",
					Owner:       "security@elastic.co",
					Description: "test",
				},
				Environment:              "production",
				RoleName:                 "foo",
				ObjectName:               "foo",
				DisableEnvironmentSuffix: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := compile(tt.args.env, tt.args.def, tt.args.specHash, tt.args.noTokenWrap, tt.args.defaultAuthEngineName)
			if (err != nil) != tt.wantErr {
				t.Errorf("compile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			specDisableEnvSuffix := tt.args.def.Spec.DisableEnvironmentSuffix
			if res.DisableEnvironmentSuffix != specDisableEnvSuffix {
				t.Errorf("compile() DisableEnvironmentSuffix = %v, want %v", res.DisableEnvironmentSuffix, specDisableEnvSuffix)
				return
			}

			switch specDisableEnvSuffix {
			case true:
				expectedObjectName := slug.Make(tt.args.def.Meta.Name)
				if res.ObjectName != expectedObjectName {
					t.Errorf("compile() ObjectName = %v, want %v", res.ObjectName, expectedObjectName)
					return
				}
			case false:
				expectedObjectName := slug.Make(fmt.Sprintf("%s-%s", tt.args.def.Meta.Name, tt.args.env))
				if res.ObjectName != expectedObjectName {
					t.Errorf("compile() ObjectName = %v, want %v", res.ObjectName, expectedObjectName)
					return
				}
			}
		})
	}
}

func Test_compile_Fuzz(t *testing.T) {
	// Making sure the descrption never panics
	for i := 0; i < 50; i++ {
		f := fuzz.New()

		// Prepare arguments
		var env string
		spec := &terraformerv1.AppRoleDefinition{
			ApiVersion: "harp.elastic.co/terraformer/v1",
			Kind:       "AppRoleDefinition",
			Meta: &terraformerv1.AppRoleDefinitionMeta{
				Name:        "foo",
				Owner:       "security-team@elastic.co",
				Description: "test",
			},
			Spec: &terraformerv1.AppRoleDefinitionSpec{
				Selector: &terraformerv1.AppRoleDefinitionSelector{},
			},
		}
		var specHash string
		var tokenWrap bool
		var authEngineName string

		// Fuzz input
		f.Fuzz(&env)
		f.Fuzz(&spec.Spec.Selector)
		f.Fuzz(&spec.Spec.Namespaces)
		f.Fuzz(&spec.Spec.Custom)
		f.Fuzz(&spec.Spec.DisableEnvironmentSuffix)
		f.Fuzz(&spec.Spec.AuthEngineName)
		f.Fuzz(&specHash)
		f.Fuzz(&tokenWrap)
		f.Fuzz(&authEngineName)

		// Execute
		compile(env, spec, specHash, tokenWrap, authEngineName)
	}
}

func Test_filterCapabilities(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty input",
			input:    []string{},
			expected: []string{"read"},
		},
		{
			name: "all allowed",
			input: []string{
				"read",
				"create",
				"list",
				"patch",
				"update",
			},
			expected: []string{
				"read",
				"create",
				"list",
				"patch",
				"update",
			},
		},
		{
			name:     "some not allowed",
			input:    []string{"read", "foo", "bar", "delete"},
			expected: []string{"read", "delete"},
		},
		{
			name:     "duplicates",
			input:    []string{"read", "read", "delete", "delete"},
			expected: []string{"read", "delete"},
		},
		{
			name:     "only not allowed",
			input:    []string{"foo", "bar"},
			expected: []string{"read"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterCapabilities(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("filterCapabilities(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func Test_compile_spec_fields(t *testing.T) {
	// Test that spec fields are the source of truth (GitOps-first)
	tests := []struct {
		name                     string
		specDisableEnvSuffix     bool
		specAuthEngineName       string
		defaultAuthEngineName    string
		expectedDisableEnvSuffix bool
		expectedAuthEngineName   string
		expectedObjectNameHasEnv bool
	}{
		{
			name:                     "spec disableEnvironmentSuffix=true",
			specDisableEnvSuffix:     true,
			specAuthEngineName:       "",
			defaultAuthEngineName:    "service",
			expectedDisableEnvSuffix: true,
			expectedAuthEngineName:   "service",
			expectedObjectNameHasEnv: false,
		},
		{
			name:                     "spec authEngineName overrides default",
			specDisableEnvSuffix:     false,
			specAuthEngineName:       "custom-approle",
			defaultAuthEngineName:    "service",
			expectedDisableEnvSuffix: false,
			expectedAuthEngineName:   "custom-approle",
			expectedObjectNameHasEnv: true,
		},
		{
			name:                     "empty spec uses defaults",
			specDisableEnvSuffix:     false,
			specAuthEngineName:       "",
			defaultAuthEngineName:    "service",
			expectedDisableEnvSuffix: false,
			expectedAuthEngineName:   "service",
			expectedObjectNameHasEnv: true,
		},
		{
			name:                     "both spec fields set",
			specDisableEnvSuffix:     true,
			specAuthEngineName:       "approle",
			defaultAuthEngineName:    "service",
			expectedDisableEnvSuffix: true,
			expectedAuthEngineName:   "approle",
			expectedObjectNameHasEnv: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := &terraformerv1.AppRoleDefinition{
				ApiVersion: "harp.elastic.co/terraformer/v1",
				Kind:       "AppRoleDefinition",
				Meta: &terraformerv1.AppRoleDefinitionMeta{
					Name:        "test-role",
					Owner:       "security@elastic.co",
					Description: "test",
				},
				Spec: &terraformerv1.AppRoleDefinitionSpec{
					Selector:                 &terraformerv1.AppRoleDefinitionSelector{},
					DisableEnvironmentSuffix: tt.specDisableEnvSuffix,
					AuthEngineName:           tt.specAuthEngineName,
				},
			}

			res, err := compile("production", def, "hash", false, tt.defaultAuthEngineName)
			if err != nil {
				t.Fatalf("compile() error = %v", err)
			}

			if res.DisableEnvironmentSuffix != tt.expectedDisableEnvSuffix {
				t.Errorf("DisableEnvironmentSuffix = %v, want %v", res.DisableEnvironmentSuffix, tt.expectedDisableEnvSuffix)
			}

			if res.AuthEngineName != tt.expectedAuthEngineName {
				t.Errorf("AuthEngineName = %v, want %v", res.AuthEngineName, tt.expectedAuthEngineName)
			}

			hasEnv := res.ObjectName == "test-role-production"
			if hasEnv != tt.expectedObjectNameHasEnv {
				t.Errorf("ObjectName = %v, expectedHasEnv = %v", res.ObjectName, tt.expectedObjectNameHasEnv)
			}
		})
	}
}
