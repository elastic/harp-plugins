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

package container

import (
	"context"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/elastic/harp/pkg/container/identity"
	"github.com/elastic/harp/pkg/container/identity/key"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/value/encryption/aead"
	"github.com/elastic/harp/pkg/sdk/value/encryption/envelope"
	"github.com/elastic/harp/pkg/tasks"
	"go.uber.org/zap"

	"github.com/elastic/harp-plugins/yubikey/pkg/value/encryption/envelope/piv"
)

// IdentityTask implements secret container identity creation task.
type IdentityTask struct {
	OutputWriter tasks.WriterProvider
	Description  string
	Serial       uint32
	Slot         uint8
}

// Run the task.
func (t *IdentityTask) Run(ctx context.Context) error {
	// Check arguments
	if t.Description == "" {
		return fmt.Errorf("description must not be blank")
	}

	// Create identity
	id, payload, err := identity.New(rand.Reader, t.Description, key.Ed25519)
	if err != nil {
		return fmt.Errorf("unable to create new identity: %w", err)
	}

	// Get PIV Manager
	manager := piv.Manager()

	// Try to open card
	card, err := manager.Open(t.Serial, t.Slot)
	if err != nil {
		return fmt.Errorf("unable to open PIV card: %w", err)
	}
	defer func() {
		if card != nil {
			if errClose := card.Close(); errClose != nil {
				log.For(ctx).Error("unable to close card", zap.Error(errClose))
			}
		}
	}()

	// Extract public key
	cardPublicKey := card.Public()
	pivCompressed := elliptic.MarshalCompressed(cardPublicKey.Curve, cardPublicKey.X, cardPublicKey.Y)

	// Identity tag
	tag := sha256.Sum256(pivCompressed)

	// Initialize envelope service
	pivService, err := piv.Service(card, func(msg string) (string, error) {
		pin, errPin := cmdutil.ReadSecret(msg, false)
		if errPin != nil {
			return "", fmt.Errorf("unable to read pin from terminal: %w", errPin)
		}

		// No error
		return pin.String(), nil
	})
	if err != nil {
		return fmt.Errorf("unable to initialize PIV service: %w", err)
	}

	// Initialize Data encryption transformer
	transformer, err := envelope.Transformer(pivService, aead.Chacha20Poly1305)
	if err != nil {
		return fmt.Errorf("unable to initialize KMS service: %w", err)
	}

	// Apply transformation
	cipherText, err := transformer.To(ctx, payload)
	if err != nil {
		return fmt.Errorf("unable to encrypt identity payload: %w", err)
	}

	// Wrap private key
	id.Private = &identity.PrivateKey{
		Encoding: fmt.Sprintf("piv:yubikey:%d:%02x:%s", t.Serial, t.Slot, base64.RawStdEncoding.EncodeToString(tag[:4])),
		Content:  base64.RawURLEncoding.EncodeToString(cipherText),
	}

	// Retrieve output writer
	writer, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve output writer handle: %w", err)
	}

	// Create identity output
	if err := json.NewEncoder(writer).Encode(id); err != nil {
		return fmt.Errorf("unable to serialize final identity: %w", err)
	}

	// No error
	return nil
}
