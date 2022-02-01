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
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/elastic/harp-plugins/cmd/harp-yubikey/pkg/value/encryption/envelope/piv"
	"github.com/elastic/harp/pkg/container/identity"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/security"
	"github.com/elastic/harp/pkg/sdk/value/encryption/aead"
	"github.com/elastic/harp/pkg/sdk/value/encryption/envelope"
	"github.com/elastic/harp/pkg/tasks"
	"go.uber.org/zap"
)

// RecoverTask implements secret container identity recovery task.
type RecoverTask struct {
	JSONReader   tasks.ReaderProvider
	OutputWriter tasks.WriterProvider
	Serial       uint32
	Slot         uint8
	JSONOutput   bool
}

// Run the task.
//nolint:funlen,gocyclo // To refactor
func (t *RecoverTask) Run(ctx context.Context) error {
	// Create input reader
	reader, err := t.JSONReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to initialize reader: %w", err)
	}

	// Extract identity
	input, err := identity.FromReader(reader)
	if err != nil {
		return fmt.Errorf("unable to read identity from reader: %w", err)
	}
	if input == nil {
		return fmt.Errorf("identity is nil")
	}

	// Parse encoding
	if !strings.HasPrefix(input.Private.Encoding, "piv:yubikey:") {
		return fmt.Errorf("this identity could not be recovered using yubikey: %s", input.Private.Encoding)
	}
	parts := strings.SplitN(strings.TrimPrefix(input.Private.Encoding, "piv:yubikey:"), ":", 3)

	// Unpack values
	serial, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return fmt.Errorf("unable to parse serial '%s' from identity: %w", parts[0], err)
	}
	slot, err := strconv.ParseUint(parts[1], 16, 8)
	if err != nil {
		return fmt.Errorf("unable to parse slot '%s' from identity: %w", parts[1], err)
	}
	tag, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("unable to extract identity tag '%s' from identity: %w", parts[2], err)
	}

	// Get PIV Manager
	manager := piv.Manager()

	// Try to open card
	card, err := manager.Open(uint32(serial), uint8(slot))
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
	idTag := sha256.Sum256(pivCompressed)

	// Compare tags
	if !security.SecureCompare(idTag[:4], tag) {
		return fmt.Errorf("invalid identity tag")
	}

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
		return fmt.Errorf("unable to initialize PIV service: %w", err)
	}

	// Try to decrypt identity
	key, err := input.Decrypt(ctx, transformer)
	if err != nil {
		return fmt.Errorf("unable to decrypt identity: %w", err)
	}

	// Check validity
	if !security.SecureCompareString(input.Public, key.X) {
		return fmt.Errorf("invalid identity, key mismatch detected")
	}

	// Get output writer
	outputWriter, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve output writer: %w", err)
	}

	// Display as json
	if t.JSONOutput {
		if err := json.NewEncoder(outputWriter).Encode(map[string]interface{}{
			"container_key": key.D,
		}); err != nil {
			return fmt.Errorf("unable to display as json: %w", err)
		}
	} else {
		// Display container key
		fmt.Fprintf(outputWriter, "Container key : %s\n", key.D)
	}

	// No error
	return nil
}
