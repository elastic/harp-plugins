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
	"testing"

	terraformerv1 "github.com/elastic/harp-plugins/cmd/harp-terraformer/api/gen/go/harp/terraformer/v1"
	fuzz "github.com/google/gofuzz"
)

func Test_compile(t *testing.T) {
	type args struct {
		env         string
		def         *terraformerv1.AppRoleDefinition
		specHash    string
		noTokenWrap bool
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
			_, err := compile(tt.args.env, tt.args.def, tt.args.specHash, tt.args.noTokenWrap)
			if (err != nil) != tt.wantErr {
				t.Errorf("compile() error = %v, wantErr %v", err, tt.wantErr)
				return
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

		// Fuzz input
		f.Fuzz(&env)
		f.Fuzz(&spec.Spec.Selector)
		f.Fuzz(&spec.Spec.Namespaces)
		f.Fuzz(&spec.Spec.Custom)
		f.Fuzz(&specHash)
		f.Fuzz(&tokenWrap)

		// Execute
		compile(env, spec, specHash, tokenWrap)
	}
}
