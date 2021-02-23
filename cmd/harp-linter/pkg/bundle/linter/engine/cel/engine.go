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

package cel

import (
	"context"
	"errors"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/ext"
	"github.com/google/cel-go/interpreter/functions"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/proto"

	"github.com/elastic/harp-plugins/cmd/harp-linter/pkg/bundle/linter/engine"
	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
)

var (
	harpPackageObjectType = decls.NewObjectType("harp.bundle.v1.Package")
	harpKVObjectType      = decls.NewObjectType("harp.bundle.v1.KV")
)

// -----------------------------------------------------------------------------

// New returns a Google CEL based linter engine.
func New(expressions []string) (engine.PackageLinter, error) {
	// Prepare CEL Environment
	env, err := cel.NewEnv(
		cel.Types(&bundlev1.Bundle{}, &bundlev1.Package{}, &bundlev1.SecretChain{}, &bundlev1.KV{}),
		cel.Declarations(
			decls.NewVar("p", harpPackageObjectType),
			decls.NewFunction("match_path",
				decls.NewInstanceOverload("match_path",
					[]*exprpb.Type{harpPackageObjectType, decls.String},
					decls.Bool,
				),
			),
			decls.NewFunction("has_secret",
				decls.NewInstanceOverload("has_secret",
					[]*exprpb.Type{harpPackageObjectType, decls.String},
					decls.Bool,
				),
			),
			decls.NewFunction("has_all_secrets",
				decls.NewInstanceOverload("has_all_secrets",
					[]*exprpb.Type{harpPackageObjectType, decls.NewListType(decls.String)},
					decls.Bool,
				),
			),
			decls.NewFunction("is_cso_compliant",
				decls.NewInstanceOverload("is_cso_compliant",
					[]*exprpb.Type{harpPackageObjectType},
					decls.Bool,
				),
			),
			decls.NewFunction("secret",
				decls.NewInstanceOverload("secret",
					[]*exprpb.Type{harpPackageObjectType, decls.String},
					harpKVObjectType,
				),
			),
			decls.NewFunction("is_base64",
				decls.NewInstanceOverload("is_base64",
					[]*exprpb.Type{harpKVObjectType},
					decls.Bool,
				),
			),
			decls.NewFunction("is_required",
				decls.NewInstanceOverload("is_required",
					[]*exprpb.Type{harpKVObjectType},
					decls.Bool,
				),
			),
			decls.NewFunction("is_url",
				decls.NewInstanceOverload("is_url",
					[]*exprpb.Type{harpKVObjectType},
					decls.Bool,
				),
			),
			decls.NewFunction("is_uuid",
				decls.NewInstanceOverload("is_uuid",
					[]*exprpb.Type{harpKVObjectType},
					decls.Bool,
				),
			),
			decls.NewFunction("is_email",
				decls.NewInstanceOverload("is_email",
					[]*exprpb.Type{harpKVObjectType},
					decls.Bool,
				),
			),
			decls.NewFunction("is_json",
				decls.NewInstanceOverload("is_json",
					[]*exprpb.Type{harpKVObjectType},
					decls.Bool,
				),
			),
		),
		ext.Strings(),
		ext.Encoders(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare CEL engine environment: %w", err)
	}

	// Registter types
	reg, err := types.NewRegistry(
		&bundlev1.KV{},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to register types: %w", err)
	}

	// Functions
	funcs := cel.Functions(
		&functions.Overload{
			Operator: "match_path",
			Binary:   celPackageMatchPath,
		},
		&functions.Overload{
			Operator: "has_secret",
			Binary:   celPackageHasSecret,
		},
		&functions.Overload{
			Operator: "has_all_secrets",
			Binary:   celPackageHasAllSecrets,
		},
		&functions.Overload{
			Operator: "is_cso_compliant",
			Unary:    celPackageIsCSOCompliant,
		},
		&functions.Overload{
			Operator: "secret",
			Binary:   celPackageGetSecret(reg),
		},
		&functions.Overload{
			Operator: "is_base64",
			Unary:    celValidatorBuilder(is.Base64),
		},
		&functions.Overload{
			Operator: "is_required",
			Unary:    celValidatorBuilder(validation.Required),
		},
		&functions.Overload{
			Operator: "is_url",
			Unary:    celValidatorBuilder(is.URL),
		},
		&functions.Overload{
			Operator: "is_uuid",
			Unary:    celValidatorBuilder(is.UUID),
		},
		&functions.Overload{
			Operator: "is_email",
			Unary:    celValidatorBuilder(is.EmailFormat),
		},
		&functions.Overload{
			Operator: "is_json",
			Unary:    celValidatorBuilder(&jsonValidator{}),
		},
	)

	// Assemble the complete ruleset
	ruleset := make([]cel.Program, 0, len(expressions))
	for _, exp := range expressions {
		// Parse expression
		parsed, issues := env.Parse(exp)
		if issues != nil && issues.Err() != nil {
			return nil, fmt.Errorf("unable to parse '%s', go error: %w", exp, issues.Err())
		}

		// Extract AST
		ast, cerr := env.Check(parsed)
		if cerr != nil && cerr.Err() != nil {
			return nil, fmt.Errorf("invalid CEL expression: %w", cerr.Err())
		}

		// request matching is a boolean operation, so we don't really know
		// what to do if the expression returns a non-boolean type
		if !proto.Equal(ast.ResultType(), decls.Bool) {
			return nil, fmt.Errorf("CEL rule engine expects return type of bool, not %s", ast.ResultType())
		}

		// Compile the program
		p, err := env.Program(ast, funcs)
		if err != nil {
			return nil, fmt.Errorf("error while creating CEL program: %w", err)
		}

		// Add to context
		ruleset = append(ruleset, p)
	}

	// Return rule engine
	return &ruleEngine{
		cel:     env,
		ruleset: ruleset,
	}, nil
}

// -----------------------------------------------------------------------------

type ruleEngine struct {
	cel     *cel.Env
	ruleset []cel.Program
}

func (re *ruleEngine) EvaluatePackage(ctx context.Context, p *bundlev1.Package) error {
	// Check arguments
	if p == nil {
		return errors.New("unable to evaluate nil package")
	}

	// Apply evaluation (implicit AND between rules)
	for _, exp := range re.ruleset {
		// Evaluate using the bundle context
		out, _, err := exp.Eval(map[string]interface{}{
			"p": p,
		})
		if err != nil {
			return fmt.Errorf("an error occurred during the rule evaluation: %w", err)
		}

		// Boolean rule returned false
		if out.Value() == false {
			return engine.ErrRuleNotValid
		}
	}

	// No error
	return nil
}
