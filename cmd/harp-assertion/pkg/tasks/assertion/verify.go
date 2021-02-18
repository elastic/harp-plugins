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
	"io/ioutil"
	"time"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"

	shtasks "github.com/elastic/harp/pkg/tasks"
)

// Verify implements assertion verification task.
type Verify struct {
	JWKS          *jose.JSONWebKeySet
	ContentReader shtasks.ReaderProvider
	OutputWriter  shtasks.WriterProvider
}

// Run the task.
func (t *Verify) Run(ctx context.Context) error {
	// Check arguments
	if t.JWKS == nil {
		return fmt.Errorf("JWKS must not be nil")
	}

	// Retrieve content reader
	reader, err := t.ContentReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to open content for read: %w", err)
	}

	// Drain input
	assertion, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("unable to read assertion from reader: %w", err)
	}
	if len(assertion) == 0 {
		return fmt.Errorf("assertion is empty")
	}

	// Parse assertion
	token, err := jwt.ParseSigned(string(assertion))
	if err != nil {
		return fmt.Errorf("invalid assertion, syntax error: %w", err)
	}

	// Check headers
	if errValidationHeaders := t.validateHeaders(token); errValidationHeaders != nil {
		return errValidationHeaders
	}

	// Iterate over JWKS
	var (
		body   map[string]interface{}
		claims jwt.Claims
	)

	valid := false
	for idx := range t.JWKS.Keys {
		// Use index for lower memory usage
		jwk := t.JWKS.Keys[idx]

		// Check signature
		if err = token.Claims(jwk, &body, &claims); err == nil {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("unable to valid assertion signature")
	}

	// Validate claims
	if errValidationClaims := t.validateClaims(&claims); errValidationClaims != nil {
		return errValidationClaims
	}

	// Allocate writer
	writer, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to allocate output writer: %w", err)
	}

	// Write payload
	if err = json.NewEncoder(writer).Encode(body); err != nil {
		return fmt.Errorf("unable to encode content as JSON: %w", err)
	}

	// No error
	return nil
}

// -----------------------------------------------------------------------------

func (t *Verify) validateHeaders(token *jwt.JSONWebToken) error {
	// Token type
	tokenType, ok := token.Headers[0].ExtraHeaders[jose.HeaderType]
	if !ok {
		return fmt.Errorf("invalid token, missing token type")
	}
	if tokenType != assertionHeaderType {
		return fmt.Errorf("invalid token, unexpected token type '%s'", tokenType)
	}

	// Embedded JWK
	var body map[string]interface{}
	if err := token.Claims(token.Headers[0].JSONWebKey, &body); err != nil {
		return fmt.Errorf("unable to validate embedded public key: %w", err)
	}

	// No error
	return nil
}

func (t *Verify) validateClaims(claims *jwt.Claims) error {
	// Check claims
	if claims.Expiry.Time().Before(time.Now()) {
		return fmt.Errorf("expired assertion")
	}
	if claims.NotBefore.Time().After(time.Now()) {
		return fmt.Errorf("assertion not useable yet")
	}
	if claims.Issuer != "harp-assertion" {
		return fmt.Errorf("invalid assertion, unexpected issuer '%s'", claims.Issuer)
	}

	// No error
	return nil
}
