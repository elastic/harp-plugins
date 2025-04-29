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
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"text/template"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"google.golang.org/protobuf/proto"

	terraformerv1 "github.com/elastic/harp-plugins/terraformer/api/gen/go/harp/terraformer/v1"
	"github.com/elastic/harp-plugins/terraformer/pkg/terraformer/hcl"
)

// -----------------------------------------------------------------------------

// Run the template generation
func Run(_ context.Context, reader io.Reader, environmentParam string, noTokenWrap bool, templateRaw string, w io.Writer) error {
	// Drain input reader
	specificationRaw, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("unable to read input specification: %w", err)
	}

	// Load YAML to Protobuf
	def, err := loadFromYAML(specificationRaw)
	if err != nil {
		return fmt.Errorf("unable to deserialize specification: %w", err)
	}

	// Validate specification
	if err = validate(def); err != nil {
		return fmt.Errorf("unable to validate specification: %w", err)
	}

	// Serialize as protobuf
	specProto, err := proto.Marshal(def)
	if err != nil {
		return fmt.Errorf("unable to prepare specification hash: %w", err)
	}

	// Calculate spechash
	specHash := sha256.Sum256(specProto)

	// Compile the definition
	m, err := compile(environmentParam, def, base64.StdEncoding.EncodeToString(specHash[:]), noTokenWrap)
	if err != nil {
		return fmt.Errorf("unable to compile specification: %w", err)
	}

	// Prepare template
	t, err := template.New("tf").Parse(templateRaw)
	if err != nil {
		return fmt.Errorf("unable ot compile template: %w", err)
	}

	// Merge with template
	var out bytes.Buffer
	err = t.Execute(&out, m)
	if err != nil {
		return fmt.Errorf("unable to merge template with specification: %w", err)
	}

	// Format output
	formatted, err := hcl.Format(&out)
	if err != nil {
		return fmt.Errorf("unable to format terraform output: %w", err)
	}

	// Write result to writer
	if _, err := fmt.Fprintf(os.Stdout, "%s", formatted); err != nil {
		return fmt.Errorf("unable to write output result: %w", err)
	}

	// No error
	return nil
}

// -----------------------------------------------------------------------------

func validate(def *terraformerv1.AppRoleDefinition) error {
	// Check argument
	if def == nil {
		return fmt.Errorf("unable to validate nil speicification")
	}

	// Header
	if def.ApiVersion != "harp.elastic.co/terraformer/v1" {
		return fmt.Errorf("apiVersion must be 'harp.elastic.co/terraformer/v1'")
	}
	if def.Kind != "AppRoleDefinition" {
		return fmt.Errorf("kind must be 'AppRoleDefinition'")
	}

	// Meta
	if def.Meta == nil {
		return fmt.Errorf("meta object is mandatory")
	}
	if def.Meta.Name == "" {
		return fmt.Errorf("meta.Name is mandatory")
	}
	if def.Meta.Owner == "" {
		return fmt.Errorf("meta.Owner is mandatory")
	}
	if err := validation.Validate(def.Meta.Owner, is.Email); err != nil {
		return fmt.Errorf("meta.Owner must be an email: %w", err)
	}
	if def.Meta.Description == "" {
		return fmt.Errorf("meta.Description is mandatory")
	}

	// Spec
	if def.Spec == nil {
		return fmt.Errorf("spec object is mandatory")
	}
	if def.Spec.Selector == nil {
		return fmt.Errorf("spec.selector object is mandatory")
	}

	return nil
}
