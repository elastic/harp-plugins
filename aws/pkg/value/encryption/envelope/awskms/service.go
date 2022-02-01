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

package awskms

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"

	"github.com/elastic/harp/pkg/sdk/types"
	"github.com/elastic/harp/pkg/sdk/value/encryption/envelope"
)

// Service returns an AWS KMS based envelope encryption service instance.
func Service(kmsClient kmsiface.KMSAPI, keyID string) (envelope.Service, error) {
	// Check arguments
	if types.IsNil(kmsClient) {
		return nil, fmt.Errorf("unable to initialize awskms service with nil client")
	}
	if keyID == "" {
		return nil, fmt.Errorf("unable to initialize awskms service with blank key id")
	}

	// Try to retreieve key information
	kid, err := getKeyInfo(kmsClient, keyID)
	if err != nil {
		return nil, err
	}

	// Return service wrapper.
	return &service{
		kmsClient: kmsClient,
		keyID:     kid,
	}, nil
}

type service struct {
	kmsClient kmsiface.KMSAPI
	keyID     string
}

// -----------------------------------------------------------------------------

func (s *service) Decrypt(ctx context.Context, encrypted []byte) ([]byte, error) {
	// Check arguments
	if types.IsNil(s.kmsClient) {
		return nil, fmt.Errorf("unable to query awskms service with nil client")
	}

	// Prepare AWS request
	input := &kms.DecryptInput{
		CiphertextBlob: encrypted,
	}

	// Call AWS KMS API
	output, err := s.kmsClient.DecryptWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt data with AWS KMS: %w", err)
	}

	// Return result
	return output.Plaintext, nil
}

func (s *service) Encrypt(ctx context.Context, cleartext []byte) ([]byte, error) {
	// Check arguments
	if types.IsNil(s.kmsClient) {
		return nil, fmt.Errorf("unable to query awskms service with nil client")
	}

	// Prepare AWS request
	input := &kms.EncryptInput{
		KeyId:     aws.String(s.keyID),
		Plaintext: cleartext,
	}

	// Call AWS KMS API
	output, err := s.kmsClient.EncryptWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt data with AWS KMS: %w", err)
	}

	// Return result
	return output.CiphertextBlob, nil
}

// -----------------------------------------------------------------------------

func getKeyInfo(kmsClient kmsiface.KMSAPI, keyARN string) (string, error) {
	// Check arguments
	if types.IsNil(kmsClient) {
		return "", fmt.Errorf("unable to initialize awskms service with nil client")
	}
	if keyARN == "" {
		return "", fmt.Errorf("unable to initialize awskms service with blank key")
	}

	// Retrieve key information from AWS
	keyInfo, err := kmsClient.DescribeKey(&kms.DescribeKeyInput{
		KeyId: aws.String(keyARN),
	})
	if err != nil {
		return "", fmt.Errorf("error fetching AWS KMS information: %w", err)
	}
	if keyInfo == nil || keyInfo.KeyMetadata == nil || keyInfo.KeyMetadata.KeyId == nil {
		return "", errors.New("no key returned")
	}

	// No error
	return aws.StringValue(keyInfo.KeyMetadata.KeyId), nil
}
