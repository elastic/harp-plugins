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
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"golang.org/x/crypto/blake2b"

	"github.com/elastic/harp-plugins/cmd/harp-aws/pkg/value/encryption/envelope/awskms"
	"github.com/elastic/harp/pkg/container/identity"
	"github.com/elastic/harp/pkg/sdk/value/encryption/aead"
	"github.com/elastic/harp/pkg/sdk/value/encryption/envelope"
	"github.com/elastic/harp/pkg/tasks"
)

// IdentityTask implements secret container identity creation task.
type IdentityTask struct {
	OutputWriter tasks.WriterProvider
	Description  string
	KeyID        string
}

// Run the task.
func (t *IdentityTask) Run(ctx context.Context) error {
	// Check arguments
	if t.Description == "" {
		return fmt.Errorf("description must not be blank")
	}

	// Create identity
	id, payload, err := identity.New(t.Description)
	if err != nil {
		return fmt.Errorf("unable to create new identity: %w", err)
	}

	// Prepare AWS KMS client
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Assemble an envelope value transformer
	awsKmsClient := kms.New(sess)

	// Initialize Key encryption transformer
	awsKMSService, err := awskms.Service(awsKmsClient, t.KeyID)
	if err != nil {
		return fmt.Errorf("unable to initialize KMS service: %w", err)
	}

	// Initialize Data encryption transformer
	transformer, err := envelope.Transformer(awsKMSService, aead.Chacha20Poly1305)
	if err != nil {
		return fmt.Errorf("unable to initialize KMS service: %w", err)
	}

	// Apply transformation
	cipherText, err := transformer.To(ctx, payload)
	if err != nil {
		return fmt.Errorf("unable to encrypt identity payload: %w", err)
	}

	// Prepare key ID
	h := blake2b.Sum256([]byte(t.KeyID))

	// Wrap private key
	id.Private = &identity.PrivateKey{
		Encoding: fmt.Sprintf("kms:aws:%s", base64.RawURLEncoding.EncodeToString(h[:])),
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
