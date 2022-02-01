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
	"crypto/ecdsa"
	"errors"
	"fmt"

	gopiv "github.com/go-piv/piv-go/piv"
)

// Prompter is used to describe PIN prompt contract.
type Prompter func(msg string) (string, error)

// Card is a PIV card abstraction.
type Card interface {
	Close() error
	Public() *ecdsa.PublicKey
	SharedKey(peer *ecdsa.PublicKey, prompt Prompter) ([]byte, error)
}

// -----------------------------------------------------------------------------

type pivCard struct {
	card   *gopiv.YubiKey
	serial uint32
	slot   gopiv.Slot
	pub    *ecdsa.PublicKey
}

func (c *pivCard) Close() error {
	if c.card == nil {
		return nil
	}
	return c.card.Close()
}

func (c *pivCard) Public() *ecdsa.PublicKey {
	return c.pub
}

func (c *pivCard) SharedKey(peer *ecdsa.PublicKey, prompt Prompter) ([]byte, error) {
	// Check arguments
	if c.card == nil {
		return nil, errors.New("card is not initialized")
	}
	if peer == nil {
		return nil, errors.New("unable to proceed with a nil peer public key")
	}

	// Extract certificate private key.
	priv, err := c.card.PrivateKey(c.slot, c.pub, gopiv.KeyAuth{
		PINPrompt: func() (string, error) {
			return prompt(fmt.Sprintf("Enter PIN for Yubikey with serial %d", c.serial))
		},
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get PIV private key handle: %w", err)
	}

	// Compute ECDH shared secret from key.
	shared, err := priv.(*gopiv.ECDSAPrivateKey).SharedKey(peer)
	if err != nil {
		return nil, fmt.Errorf("PIV ECDHE error: %w", err)
	}

	// No error
	return shared, nil
}
