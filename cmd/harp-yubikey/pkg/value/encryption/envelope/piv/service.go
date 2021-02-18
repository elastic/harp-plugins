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

package piv

import (
	"context"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"github.com/elastic/harp/pkg/sdk/security"
	"github.com/elastic/harp/pkg/sdk/value/encryption/envelope"
	"github.com/fxamacker/cbor/v2"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

// Service returns an PIV based envelope encryption service instance.
func Service(card Card, prompt Prompter) (envelope.Service, error) {
	// Return service wrapper.
	return &service{
		card:   card,
		prompt: prompt,
	}, nil
}

const wrapLabel = "harp.elastic.co/v1/piv"

type service struct {
	card   Card
	prompt Prompter
}

// -----------------------------------------------------------------------------

func (s *service) Decrypt(ctx context.Context, encrypted []byte) ([]byte, error) {
	// Extract envelope
	var r response
	if err := cbor.Unmarshal(encrypted, &r); err != nil {
		return nil, fmt.Errorf("unable to decode envelope: %w", err)
	}

	// Extract public keys
	pivPublicKey := s.card.Public()
	pivCompressed := elliptic.MarshalCompressed(pivPublicKey.Curve, pivPublicKey.X, pivPublicKey.Y)

	// Identity tag
	tag := sha256.Sum256(pivCompressed)

	// Compare tag
	if !security.SecureCompare(tag[:4], r.Tag) {
		return nil, errors.New("invalid identity tag")
	}

	// Extract ephemeral public key
	x, y := elliptic.UnmarshalCompressed(pivPublicKey.Curve, r.EphCompressedPublic)
	if x == nil {
		return nil, errors.New("cannot unmarshal ephemeral public key")
	}
	ephPub := &ecdsa.PublicKey{
		Curve: pivPublicKey.Curve,
		X:     x,
		Y:     y,
	}

	// Compute shared secret
	sharedSecret, err := s.card.SharedKey(ephPub, s.prompt)
	if err != nil {
		return nil, fmt.Errorf("unable to compute shared secret: %w", err)
	}

	// Derive AEAD cipher from parameters
	aead, err := s.deriveAEAD(r.EphCompressedPublic, pivCompressed, sharedSecret, chacha20poly1305.KeySize)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize aead: %w", err)
	}

	// Decrypt
	// Use fixed nonce to save space and also sharedsecret is derived from ephemeral
	// key that act as nonce.
	nonce := make([]byte, chacha20poly1305.NonceSize)
	clearText, err := aead.Open(nil, nonce, r.Payload, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt payload: %w", err)
	}

	// No error
	return clearText, nil
}

func (s *service) Encrypt(ctx context.Context, cleartext []byte) ([]byte, error) {
	// Extract public keys
	pivPublicKey := s.card.Public()
	pivCompressed := elliptic.MarshalCompressed(pivPublicKey.Curve, pivPublicKey.X, pivPublicKey.Y)

	// Identity tag
	tag := sha256.Sum256(pivCompressed)

	// Generate ephemeral key
	eph, err := ecdsa.GenerateKey(pivPublicKey.Curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("unable to generate ephemeral key pair: %w", err)
	}
	ephCompressed := elliptic.MarshalCompressed(eph.Curve, eph.PublicKey.X, eph.PublicKey.Y)

	// ECDH shared secret between ephemeral key and yubikey
	sharedSecretNum, _ := eph.PublicKey.ScalarMult(s.card.Public().X, s.card.Public().Y, eph.D.Bytes())
	sharedSecret := sharedSecretNum.Bytes()

	// Derive AEAD cipher from parameters
	aead, err := s.deriveAEAD(ephCompressed, pivCompressed, sharedSecret, chacha20poly1305.KeySize)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize aead: %w", err)
	}

	// Encrypt
	// Use fixed nonce to save space and also sharedsecret is derived from ephemeral
	// key that act as nonce.
	nonce := make([]byte, chacha20poly1305.NonceSize)

	// Return encrypted content
	body, err := cbor.Marshal(response{
		EphCompressedPublic: ephCompressed,
		Tag:                 tag[:4],
		Payload:             aead.Seal(nil, nonce, cleartext, nil),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to encode envelope: %w", err)
	}

	// No error
	return body, nil
}

// -----------------------------------------------------------------------------

func (s *service) deriveAEAD(ephPub, pivPub, sharedSecret []byte, keySize int) (cipher.AEAD, error) {
	// EphemeralPubKey || YubikeyPubKey
	salt := make([]byte, 0, len(ephPub)+len(pivPub))
	salt = append(salt, ephPub...)
	salt = append(salt, pivPub...)

	// Stretch sharedsecret to required size.
	h := hkdf.New(sha256.New, sharedSecret, salt, []byte(wrapLabel))
	wrappingKey := make([]byte, keySize)
	if _, errRand := io.ReadFull(h, wrappingKey); errRand != nil {
		return nil, fmt.Errorf("unable to generate wrapping key: %w", errRand)
	}

	// Prepare AEAD encryption
	aead, err := chacha20poly1305.New(wrappingKey)
	if err != nil {
		return nil, fmt.Errorf("unabe to initialize AEAD: %w", err)
	}

	// No error
	return aead, nil
}
