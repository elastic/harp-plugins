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

package to

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	cloudsession "github.com/elastic/harp/pkg/cloud/aws/session"
	"github.com/elastic/harp/pkg/container"
	"github.com/elastic/harp/pkg/tasks"
)

// S3UploadTask implements secret container uploading to S3 task.
type S3UploadTask struct {
	ContainerReader    tasks.ReaderProvider
	OutputWriter       tasks.WriterProvider
	Endpoint           string
	Region             string
	BucketName         string
	ObjectKey          string
	Profile            string
	AccessKeyID        string
	SecretAccessKey    string
	SessionToken       string
	DisableSSL         bool
	S3ForcePathStyle   bool
	IgnoreEnvCreds     bool
	IgnoreEC2RoleCreds bool
	IgnoreConfigCreds  bool
	JSONOutput         bool
}

// Run the task.
func (t *S3UploadTask) Run(ctx context.Context) error {
	// Create input reader
	reader, err := t.ContainerReader(ctx)
	if err != nil {
		return fmt.Errorf("unable to initialize reader: %w", err)
	}

	// Drain all content
	payload, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("unable to drain input reader: %w", err)
	}

	// Extract container
	c, err := container.Load(bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("unable to load container from input: %w", err)
	}
	if c == nil {
		return fmt.Errorf("container is nil")
	}

	// Retrieve aws session
	sess := session.Must(cloudsession.NewSession(&cloudsession.Options{
		AccessKeyID:        t.AccessKeyID,
		BucketName:         t.BucketName,
		DisableSSL:         t.DisableSSL,
		Endpoint:           t.Endpoint,
		IgnoreConfigCreds:  t.IgnoreConfigCreds,
		IgnoreEC2RoleCreds: t.IgnoreEC2RoleCreds,
		IgnoreEnvCreds:     t.IgnoreEnvCreds,
		ObjectKey:          t.ObjectKey,
		Profile:            t.Profile,
		Region:             t.Region,
		S3ForcePathStyle:   t.S3ForcePathStyle,
		SecretAccessKey:    t.SecretAccessKey,
		SessionToken:       t.SessionToken,
	}))

	// Create uploader service
	uploader := s3manager.NewUploader(sess)

	// Prepare input
	input := &s3manager.UploadInput{
		Bucket: aws.String(t.BucketName),
		Key:    aws.String(t.ObjectKey),
		Body:   bytes.NewReader(payload),
	}

	// Send the request
	up, errUpload := uploader.UploadWithContext(ctx, input)
	if errUpload != nil {
		return fmt.Errorf("unable to upload secret container to S3 bucket: %w", errUpload)
	}

	// Get output writer
	outputWriter, err := t.OutputWriter(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve output writer: %w", err)
	}

	// Display as json
	if t.JSONOutput {
		if err := json.NewEncoder(outputWriter).Encode(up); err != nil {
			return fmt.Errorf("unable to display as json: %w", err)
		}
	} else {
		// Display container key
		fmt.Fprintf(outputWriter, "Container successfully uploaded to: %s\n", up.Location)
	}

	// No error
	return nil
}
