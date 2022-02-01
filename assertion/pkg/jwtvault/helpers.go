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

package jwtvault

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"path"

	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/square/go-jose.v2"
)

// -----------------------------------------------------------------------------

type keyResponse struct {
	Type          string                `json:"type" mapstructure:"type"`
	LatestVersion float64               `json:"latest_version" mapstructure:"latest_version"`
	Keys          map[string]keyVersion `json:"keys" mapstructure:"keys"`
}

type keyVersion struct {
	PublicKey string `json:"public_key" mapstructure:"public_key"`
}

// -----------------------------------------------------------------------------

// GetPublicKey returns parsed public key
func GetPublicKey(vaultClient *api.Client, transitPath, keyName string) (publicKey interface{}, version uint, err error) {
	// Check arguments
	if vaultClient == nil {
		return nil, 0, fmt.Errorf("vault client must not be nil")
	}
	if transitPath == "" {
		return nil, 0, fmt.Errorf("transit path path must not be blank")
	}
	if keyName == "" {
		return nil, 0, fmt.Errorf("key name must not be blank")
	}

	// Retrieve transit key
	d, err := vaultClient.Logical().Read(path.Join(transitPath, "keys", keyName))
	if err != nil {
		return nil, 0, fmt.Errorf("unable to retrieve key details: %w", err)
	}
	if d == nil {
		return nil, 0, fmt.Errorf("returned key details are nil")
	}

	// Decode data
	var transitKey keyResponse
	if err = mapstructure.Decode(d.Data, &transitKey); err != nil {
		return nil, 0, fmt.Errorf("unable to decode key response: %w", err)
	}

	// Get latest version
	latestVersion, ok := transitKey.Keys[fmt.Sprintf("%d", uint(transitKey.LatestVersion))]
	if !ok {
		return nil, 0, fmt.Errorf("unable to retrieve transit key version '%f'", transitKey.LatestVersion)
	}

	// Decode PEM
	block, _ := pem.Decode([]byte(latestVersion.PublicKey))
	if block == nil {
		return nil, 0, fmt.Errorf("unable to decode public key PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to decode publiv key: %w", err)
	}

	// No error
	return pub, uint(transitKey.LatestVersion), nil
}

// JWKS extracts the public key set from a vault transit key.
func JWKS(vaultClient *api.Client, transitPath, keyName string) (*jose.JSONWebKeySet, error) {
	// Check arguments
	if vaultClient == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if transitPath == "" {
		return nil, fmt.Errorf("transit path path must not be blank")
	}
	if keyName == "" {
		return nil, fmt.Errorf("key name must not be blank")
	}

	// Retrieve transit key
	tk, err := vaultClient.Logical().Read(path.Join(transitPath, "keys", keyName))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve key details: %w", err)
	}
	if tk == nil {
		return nil, fmt.Errorf("returned key details are nil")
	}

	// Decode data
	var transitKey keyResponse
	if err = mapstructure.Decode(tk.Data, &transitKey); err != nil {
		return nil, fmt.Errorf("unable to decode transit key response: %w", err)
	}

	// Prepare key set
	jwks := &jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{},
	}

	// Iterate over all keys
	for kid, keyVersion := range transitKey.Keys {
		// Decode PEM
		block, _ := pem.Decode([]byte(keyVersion.PublicKey))
		if block == nil {
			return nil, fmt.Errorf("unable to decode public key PEM block")
		}

		// Parse key
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("unable to decode publiv key: %w", err)
		}

		// Prepare JWK
		jwks.Keys = append(jwks.Keys, jose.JSONWebKey{
			KeyID: fmt.Sprintf("vault:%s:%s:v%s", transitPath, keyName, kid),
			Key:   pub,
		})
	}

	// No error
	return jwks, nil
}
