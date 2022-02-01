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

	"github.com/elastic/harp/pkg/sdk/log"
	gopiv "github.com/go-piv/piv-go/piv"
	"go.uber.org/zap"
)

// Opener describes PIV manager card opener contract.
type Opener interface {
	Open(serial uint32, slot uint8) (Card, error)
}

// -----------------------------------------------------------------------------

const (
	pivOrganization = "harp-plugin-yubikey"
)

// Manager returns a PIV manager instance to retrieve Card info.
func Manager() Opener {
	return &pivOpener{}
}

type pivOpener struct{}

// -----------------------------------------------------------------------------

func (o *pivOpener) Open(serial uint32, slot uint8) (Card, error) {
	// Get a retired slot.
	pivSlot, ok := gopiv.RetiredKeyManagementSlot(uint32(slot))
	if !ok {
		return nil, fmt.Errorf("unrecognized slot: %02x", slot)
	}

	// Retrieve all cards
	cards, err := gopiv.Cards()
	if err != nil {
		return nil, fmt.Errorf("cannot list PIV cards: %w", err)
	}

	for _, name := range cards {
		// Try to open card by name
		card, err := o.tryOpen(name, serial)
		if err != nil {
			log.Bg().Debug("ignoring card", zap.Error(err), zap.String("name", name), zap.Uint32("serial", serial))
			continue
		}

		// Retrieve certificate
		cert, err := card.Certificate(pivSlot)
		if err != nil {
			if errors.Is(err, gopiv.ErrNotFound) {
				log.Bg().Debug("ignoring card without certificate", zap.Error(err), zap.String("name", name), zap.Uint32("serial", serial), zap.String("slot", fmt.Sprintf("%02x", slot)))
			} else {
				log.Bg().Debug("communication error", zap.Error(err), zap.String("name", name), zap.Uint32("serial", serial), zap.String("slot", fmt.Sprintf("%02x", slot)))
			}
			continue
		}

		// Filter on required organization
		orgs := cert.Subject.Organization
		if len(orgs) != 1 || orgs[0] != pivOrganization {
			log.Bg().Debug("ignoring card with wrong organization", zap.Error(err), zap.String("name", name), zap.Uint32("serial", serial), zap.String("slot", fmt.Sprintf("%02x", slot)))
			continue
		}

		// Wrap card and return
		return &pivCard{
			card:   card,
			serial: serial,
			slot:   pivSlot,
			pub:    cert.PublicKey.(*ecdsa.PublicKey),
		}, nil
	}

	return nil, errors.New("card not found")
}

// -----------------------------------------------------------------------------

func (o *pivOpener) tryOpen(name string, wantSerial uint32) (*gopiv.YubiKey, error) {
	// Open card
	card, err := gopiv.Open(name)
	if err != nil {
		return nil, fmt.Errorf("cannot open PIV card: %w", err)
	}

	// Read serial number
	gotSerial, err := card.Serial()
	if err != nil {
		return nil, fmt.Errorf("cannot get PIV card serial: %w", err)
	}

	// Compare with given
	if gotSerial != wantSerial {
		return nil, fmt.Errorf("unwanted serial: %08x", gotSerial)
	}

	// No error
	return card, nil
}
