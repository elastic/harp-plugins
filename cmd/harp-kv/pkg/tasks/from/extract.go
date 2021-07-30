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

package from

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/elastic/harp-plugins/cmd/harp-kv/pkg/kv"
	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle"
	"github.com/elastic/harp/pkg/bundle/secret"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/tasks"
)

type ExtractTask struct {
	_               struct{}
	ContainerWriter tasks.WriterProvider
	BasePaths       []string
	Store           kv.Store
}

func (t *ExtractTask) Run(ctx context.Context) error {
	// Prepare a bundle
	b := &bundlev1.Bundle{
		Packages: make([]*bundlev1.Package, 0),
	}

	// For each base path
	for _, basePath := range t.BasePaths {
		// List recusively items
		items, err := t.Store.List(ctx, basePath)
		if err != nil {
			return fmt.Errorf("unable to extract key from store: %w", err)
		}

		// Prepare a package using each item
		for _, item := range items {
			// Prepare secret list
			chain := &bundlev1.SecretChain{
				Version:         uint32(0),
				Data:            make([]*bundlev1.KV, 0),
				NextVersion:     nil,
				PreviousVersion: nil,
			}

			// Try to extract value as a json map
			var secretData map[string]interface{}
			errJSON := json.Unmarshal(item.Value, &secretData)
			if errJSON != nil {
				log.For(ctx).Warn("data could not be decoded as json", zap.Error(errJSON))

				// Create an arbitrary secret key
				secretKey := strings.TrimPrefix(item.Key, kv.GetDirectory(item.Key))

				// Pack secret value
				s, errPack := t.packSecret(secretKey, item.Value)
				if errPack != nil {
					return fmt.Errorf("unable to pack secret value for path '%s' with key '%s' : %w", item.Key, secretKey, errPack)
				}

				// Add secret to package
				chain.Data = append(chain.Data, s)
			} else {
				// Iterate over secret bundle
				for k, v := range secretData {
					// Pack secret value
					s, errPack := t.packSecret(k, v)
					if errPack != nil {
						return fmt.Errorf("unable to pack secret value for path '%s' with key '%s' : %w", item.Key, k, errPack)
					}

					// Add secret to package
					chain.Data = append(chain.Data, s)
				}
			}

			// Prepare the secret package
			pack := &bundlev1.Package{
				Labels:      map[string]string{},
				Annotations: map[string]string{},
				Name:        item.Key,
				Secrets:     chain,
			}

			// Add to bundle
			b.Packages = append(b.Packages, pack)
		}
	}

	// Create container
	writer, err := t.ContainerWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to initialize container writer: %w", err)
	}

	// Dump bundle
	if err = bundle.ToContainerWriter(writer, b); err != nil {
		return fmt.Errorf("unable to produce exported bundle: %w", err)
	}

	return nil
}

// -----------------------------------------------------------------------------
func (t *ExtractTask) packSecret(key string, value interface{}) (*bundlev1.KV, error) {
	// Pack secret value
	payload, err := secret.Pack(value)
	if err != nil {
		return nil, fmt.Errorf("unable to pack secret '%s': %w", key, err)
	}

	// Build the secret object
	return &bundlev1.KV{
		Key:   key,
		Type:  fmt.Sprintf("%T", value),
		Value: payload,
	}, nil
}
