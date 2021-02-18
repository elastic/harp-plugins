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
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"golang.org/x/crypto/blake2b"

	"github.com/elastic/harp-plugins/cmd/harp-aws/pkg/value/encryption/envelope/awskms"
	"github.com/elastic/harp/pkg/container/identity"
	"github.com/elastic/harp/pkg/sdk/security"
	"github.com/elastic/harp/pkg/sdk/value/encryption/aead"
	"github.com/elastic/harp/pkg/sdk/value/encryption/envelope"
	"github.com/elastic/harp/pkg/tasks"
)

// RecoverTask implements secret container identity recovery task.
type RecoverTask struct {
	JSONReader   tasks.ReaderProvider
	OutputWriter tasks.WriterProvider
	Description  string
	KeyID        string
	JSONOutput   bool
}

// Run the task.
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

	// Prepare key ID
	h := blake2b.Sum256([]byte(t.KeyID))
	if !strings.HasPrefix(input.Private.Encoding, fmt.Sprintf("kms:aws:%s", base64.RawURLEncoding.EncodeToString(h[:]))) {
		return fmt.Errorf("invalid identity encoding or not handled by this tool or KMS key not matching")
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
