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

package hcl

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

// Format HCL using HCLv2 spec
// https://github.com/GoogleCloudPlatform/terraformer/blob/master/terraform_utils/hcl.go
func Format(in fmt.Stringer) ([]byte, error) {
	s := in.String()

	// ...but leave whitespace between resources
	s = strings.ReplaceAll(s, "}\nresource", "}\n\nresource")

	// Workaround HCL insanity kubernetes/kops#6359: quotes are _not_ escaped in quotes (huh?)
	// This hits the file function
	s = strings.ReplaceAll(s, "(\\\"", "(\"")
	s = strings.ReplaceAll(s, "\\\")", "\")")

	// Apply Terraform style (alignment etc.)
	var err error
	formatted := hclwrite.Format([]byte(s))
	if err != nil {
		log.Println("Invalid HCL follows:")
		for i, line := range strings.Split(s, "\n") {
			fmt.Fprintf(os.Stdout, "%4d|\t%s\n", i+1, line)
		}
		return nil, fmt.Errorf("error formatting HCL: %w", err)
	}

	return formatted, nil
}
