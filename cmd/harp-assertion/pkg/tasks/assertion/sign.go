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

package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dchest/uniuri"
	"github.com/hashicorp/vault/api"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"

	"github.com/elastic/harp-plugins/cmd/harp-assertion/pkg/jwtvault"
	shtasks "github.com/elastic/harp/pkg/tasks"
)

const (
	assertionHeaderType = "harp-assertion+jwt"
)

// Sign implements attestation signing task.
type Sign struct {
	VaultClient    *api.Client
	TransitKeyName string
	Subject        string
	Audiences      []string
	Expiration     time.Duration
	ContentReader  shtasks.ReaderProvider
	OutputWriter   shtasks.WriterProvider
}

// Run the task.
func (t *Sign) Run(ctx context.Context) error {
	// Check arguments
	if t.VaultClient == nil {
		return fmt.Errorf("vault client must not be nil")
	}
	if t.TransitKeyName == "" {
		return fmt.Errorf("transit key name is mandatory")
	}
	if t.Subject == "" {
		return fmt.Errorf("subject must not be blank")
	}
	if len(t.Audiences) == 0 {
		return fmt.Errorf("audiences must not be empty")
	}
	if t.Expiration < 1*time.Minute {
		t.Expiration = 1 * time.Minute
	}
	if t.Expiration > 15*time.Minute {
		t.Expiration = 15 * time.Minute
	}

	// Retrieve content reader
	reader, err := t.ContentReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to open content for read: %w", err)
	}

	// Check input as json
	var body map[string]interface{}
	if err = json.NewDecoder(reader).Decode(&body); err != nil {
		return fmt.Errorf("unable to decode input content as JSON: %w", err)
	}

	// Retrieve public key
	pubKey, version, err := jwtvault.GetPublicKey(t.VaultClient, "assertions", t.TransitKeyName)
	if err != nil {
		return fmt.Errorf("unable to retrieve public key: %w", err)
	}

	// Allocate opaque signer
	opaqueSigner := jwtvault.Signer(t.VaultClient, "assertions", t.TransitKeyName, pubKey)

	// Get Public key
	alg := jose.SignatureAlgorithm(opaqueSigner.Public().Algorithm)
	signer, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: alg,
			Key: &jose.JSONWebKey{
				Algorithm: string(alg),
				Key:       opaqueSigner,
				KeyID:     fmt.Sprintf("vault:assertions:%s:v%d", t.TransitKeyName, version),
				Use:       "sig",
			},
		},
		&jose.SignerOptions{
			EmbedJWK: true,
			ExtraHeaders: map[jose.HeaderKey]interface{}{
				jose.HeaderType: assertionHeaderType,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("unable to allocate token signer: %w", err)
	}

	// Generate token
	now := time.Now()
	assertion, err := jwt.Signed(signer).Claims(body).Claims(&jwt.Claims{
		ID:        uniuri.NewLen(16),
		Expiry:    jwt.NewNumericDate(now.Add(t.Expiration)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now.Add(-1 * time.Second)),
		Issuer:    "harp-assertion",
		Subject:   t.Subject,
		Audience:  jwt.Audience(t.Audiences),
	}).CompactSerialize()
	if err != nil {
		return fmt.Errorf("unable to sign final assertion: %w", err)
	}

	// Allocate writer
	writer, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to initialize output writer: %w", err)
	}

	// Dump to writer
	if _, err := fmt.Fprintf(writer, "%s", assertion); err != nil {
		return fmt.Errorf("unable to write assertion to writer: %w", err)
	}

	// No error
	return nil
}
