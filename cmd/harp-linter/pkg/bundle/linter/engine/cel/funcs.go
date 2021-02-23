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
	"reflect"
	"strings"

	"github.com/gobwas/glob"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	csov1 "github.com/elastic/harp/pkg/cso/v1"
	htypes "github.com/elastic/harp/pkg/sdk/types"
)

func celPackageMatchPath(lhs, rhs ref.Val) ref.Val {
	x, _ := lhs.ConvertToNative(reflect.TypeOf(&bundlev1.Package{}))
	p := x.(*bundlev1.Package)

	pathTyped := rhs.(types.String)
	path := pathTyped.Value().(string)

	return types.Bool(glob.MustCompile(path).Match(p.Name))
}

func celPackageHasSecret(lhs, rhs ref.Val) ref.Val {
	x, _ := lhs.ConvertToNative(reflect.TypeOf(&bundlev1.Package{}))
	p := x.(*bundlev1.Package)
	secretTyped := rhs.(types.String)
	secretName := secretTyped.Value().(string)

	// No secret data
	if p.Secrets == nil || p.Secrets.Data == nil || len(p.Secrets.Data) == 0 {
		return types.Bool(false)
	}

	// Look for secret name
	for _, k := range p.Secrets.Data {
		if strings.EqualFold(k.Key, secretName) {
			return types.Bool(true)
		}
	}

	return types.Bool(false)
}

func celPackageHasAllSecrets(lhs, rhs ref.Val) ref.Val {
	x, _ := lhs.ConvertToNative(reflect.TypeOf(&bundlev1.Package{}))
	p := x.(*bundlev1.Package)
	secretsTyped, _ := rhs.ConvertToNative(reflect.TypeOf([]string{}))
	secretNames := secretsTyped.([]string)

	// No secret data
	if p.Secrets == nil || p.Secrets.Data == nil || len(p.Secrets.Data) == 0 {
		return types.Bool(false)
	}

	sa := htypes.StringArray(secretNames)

	secretMap := map[string]*bundlev1.KV{}
	for _, k := range p.Secrets.Data {
		if !sa.Contains(k.Key) {
			return types.Bool(false)
		}
		secretMap[k.Key] = k
	}

	// Look for secret name
	for _, k := range secretNames {
		if _, ok := secretMap[k]; !ok {
			return types.Bool(false)
		}
	}

	return types.Bool(true)
}

func celPackageIsCSOCompliant(lhs ref.Val) ref.Val {
	x, _ := lhs.ConvertToNative(reflect.TypeOf(&bundlev1.Package{}))
	p := x.(*bundlev1.Package)

	if err := csov1.Validate(p.Name); err != nil {
		return types.Bool(false)
	}

	return types.Bool(true)
}
