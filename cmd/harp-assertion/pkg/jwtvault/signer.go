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
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"path"
	"strings"

	"github.com/hashicorp/vault/api"
	"gopkg.in/square/go-jose.v2"
)

// Signer returns an OpaqueSigner implementation where private key is handled by
// Vault.
func Signer(vaultClient *api.Client, transitPath, keyName string, publicKey interface{}) jose.OpaqueSigner {
	return &opaqueSigner{
		vaultClient: vaultClient,
		transitPath: transitPath,
		keyName:     keyName,
		publicKey:   publicKey,
	}
}

// -----------------------------------------------------------------------------

type opaqueSigner struct {
	vaultClient *api.Client
	transitPath string
	keyName     string
	publicKey   interface{}
}

func (o *opaqueSigner) Public() *jose.JSONWebKey {
	// Prepare JWK
	k := &jose.JSONWebKey{
		Algorithm: string(jose.ES384),
		Key:       o.publicKey,
	}

	// Return JWK
	return k
}

func (o *opaqueSigner) Algs() []jose.SignatureAlgorithm {
	return []jose.SignatureAlgorithm{
		jose.ES384,
	}
}

func (o *opaqueSigner) SignPayload(payload []byte, alg jose.SignatureAlgorithm) ([]byte, error) {
	// Check arguments
	if o.vaultClient == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if o.transitPath == "" {
		return nil, fmt.Errorf("transit path path must not be blank")
	}
	if o.keyName == "" {
		return nil, fmt.Errorf("key name must not be blank")
	}

	// Compute sha512/384 hash
	h := sha512.New384()
	if _, err := h.Write(payload); err != nil {
		return nil, fmt.Errorf("unable to compute sha512/384 hash of payload: %w", err)
	}

	// Sign with transit key
	d, err := o.vaultClient.Logical().Write(path.Join(o.transitPath, "sign", o.keyName), map[string]interface{}{
		"prehashed":            true,  // Send hash only
		"marshaling_algorithm": "jws", // Force JWS Encoding
		"input":                base64.StdEncoding.EncodeToString(h.Sum(nil)),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to sign token: %w", err)
	}
	if d == nil {
		return nil, fmt.Errorf("returned signature is nil")
	}

	// Check if response have a signature
	sig, sigOk := d.Data["signature"]
	if !sigOk {
		return nil, fmt.Errorf("signature not found is response")
	}

	// Clean signature
	cleanSig := sig.(string)
	// vault:v1:<base64url>
	sigParts := strings.SplitN(cleanSig, ":", 3)

	// Decode signature
	signatureBytes, err := base64.RawURLEncoding.DecodeString(sigParts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid signature encoding: %w", err)
	}

	// No error
	return signatureBytes, nil
}
