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
	"encoding/base64"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
)

// -----------------------------------------------------------------------------

func TestService(t *testing.T) {
	type args struct {
		kmsClient kmsiface.KMSAPI
		keyID     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "key blank",
			args: args{
				kmsClient: &mockClient{},
				keyID:     "",
			},
			wantErr: true,
		},
		{
			name: "key not found",
			args: args{
				kmsClient: &mockClient{},
				keyID:     "not-found",
			},
			wantErr: true,
		},
		{
			name: "valid",
			args: args{
				kmsClient: &mockClient{
					keyID: aws.String("1234567890"),
				},
				keyID: "1234567890",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Service(tt.args.kmsClient, tt.args.keyID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_service_Decrypt(t *testing.T) {
	type fields struct {
		kmsClient kmsiface.KMSAPI
		keyID     string
	}
	type args struct {
		ctx       context.Context
		encrypted []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "empty",
			fields: fields{
				kmsClient: &mockClient{
					keyID: aws.String("1234567890"),
				},
				keyID: "1234567890",
			},
			args: args{
				ctx:       context.Background(),
				encrypted: []byte{},
			},
			want: []byte{},
		},
		{
			name: "valid",
			fields: fields{
				kmsClient: &mockClient{
					keyID: aws.String("1234567890"),
				},
				keyID: "1234567890",
			},
			args: args{
				ctx:       context.Background(),
				encrypted: []byte("bXktd29uZGVyZnVsLXNlY3JldA=="),
			},
			want: []byte("my-wonderful-secret"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				kmsClient: tt.fields.kmsClient,
				keyID:     tt.fields.keyID,
			}
			got, err := s.Decrypt(tt.args.ctx, tt.args.encrypted)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.Decrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_Encrypt(t *testing.T) {
	type fields struct {
		kmsClient kmsiface.KMSAPI
		keyID     string
	}
	type args struct {
		ctx       context.Context
		cleartext []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "nil",
			wantErr: true,
		},
		{
			name: "empty",
			fields: fields{
				kmsClient: &mockClient{
					keyID: aws.String("1234567890"),
				},
				keyID: "1234567890",
			},
			args: args{
				ctx:       context.Background(),
				cleartext: []byte{},
			},
			want: []byte{},
		},
		{
			name: "valid",
			fields: fields{
				kmsClient: &mockClient{
					keyID: aws.String("1234567890"),
				},
				keyID: "1234567890",
			},
			args: args{
				ctx:       context.Background(),
				cleartext: []byte("my-wonderful-secret"),
			},
			want: []byte("bXktd29uZGVyZnVsLXNlY3JldA=="),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				kmsClient: tt.fields.kmsClient,
				keyID:     tt.fields.keyID,
			}
			got, err := s.Encrypt(tt.args.ctx, tt.args.cleartext)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.Encrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}

// -----------------------------------------------------------------------------

type mockClient struct {
	kmsiface.KMSAPI
	keyID *string
}

// Encrypt is a mocked call that returns a base64 encoded string.
func (m *mockClient) EncryptWithContext(ctx aws.Context, input *kms.EncryptInput, opts ...request.Option) (*kms.EncryptOutput, error) {
	m.keyID = input.KeyId

	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(input.Plaintext)))
	base64.StdEncoding.Encode(encoded, input.Plaintext)

	return &kms.EncryptOutput{
		CiphertextBlob: encoded,
		KeyId:          input.KeyId,
	}, nil
}

// Decrypt is a mocked call that returns a decoded base64 string.
func (m *mockClient) DecryptWithContext(ctx aws.Context, input *kms.DecryptInput, opts ...request.Option) (*kms.DecryptOutput, error) {
	decLen := base64.StdEncoding.DecodedLen(len(input.CiphertextBlob))
	decoded := make([]byte, decLen)
	len, err := base64.StdEncoding.Decode(decoded, input.CiphertextBlob)
	if err != nil {
		return nil, err
	}

	if len < decLen {
		decoded = decoded[:len]
	}

	return &kms.DecryptOutput{
		KeyId:     m.keyID,
		Plaintext: decoded,
	}, nil
}

// DescribeKey is a mocked call that returns the keyID.
func (m *mockClient) DescribeKey(input *kms.DescribeKeyInput) (*kms.DescribeKeyOutput, error) {
	if m.keyID == nil {
		return nil, awserr.New(kms.ErrCodeNotFoundException, "key not found", nil)
	}

	return &kms.DescribeKeyOutput{
		KeyMetadata: &kms.KeyMetadata{
			KeyId: m.keyID,
		},
	}, nil
}
