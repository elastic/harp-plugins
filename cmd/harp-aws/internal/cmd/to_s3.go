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

package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/elastic/harp-plugins/cmd/harp-aws/pkg/tasks/to"
	"github.com/elastic/harp/pkg/sdk/cmdutil"
	"github.com/elastic/harp/pkg/sdk/log"
)

// -----------------------------------------------------------------------------
type toS3Params struct {
	containerPath      string
	endpoint           string
	region             string
	accessKeyID        string
	secretAccessKey    string
	sessionToken       string
	bucketName         string
	objectKey          string
	profile            string
	disableSSL         bool
	s3ForcePathStyle   bool
	ignoreConfigCreds  bool
	ignoreEC2RoleCreds bool
	ignoreEnvCreds     bool
	jsonOutput         bool
}

var toS3Cmd = func() *cobra.Command {
	params := toS3Params{}

	cmd := &cobra.Command{
		Use:   "s3",
		Short: "Push a sealed container to S3",
		Run: func(cmd *cobra.Command, _ []string) {
			// Initialize logger and context
			ctx, cancel := cmdutil.Context(cmd.Context(), "harp-aws-to-s3", conf.Debug.Enable, conf.Instrumentation.Logs.Level)
			defer cancel()

			// Override object key with filename if specified
			objectKey := params.objectKey
			if params.objectKey == "" {
				// Unspecificed or stdin convention
				if !(params.containerPath == "" || params.containerPath == "-") {
					_, file := filepath.Split(params.containerPath)
					params.objectKey = file
				}
				if params.objectKey == "" {
					log.For(ctx).Fatal("object-key flag must be specified for stdin content")
				}
			}

			// Prepare task
			t := &to.S3UploadTask{
				ContainerReader:    cmdutil.FileReader(params.containerPath),
				OutputWriter:       cmdutil.StdoutWriter(),
				Endpoint:           params.endpoint,
				Region:             params.region,
				BucketName:         params.bucketName,
				ObjectKey:          objectKey,
				Profile:            params.profile,
				AccessKeyID:        params.accessKeyID,
				SecretAccessKey:    params.secretAccessKey,
				SessionToken:       params.sessionToken,
				DisableSSL:         params.disableSSL,
				S3ForcePathStyle:   params.s3ForcePathStyle,
				IgnoreConfigCreds:  params.ignoreConfigCreds,
				IgnoreEC2RoleCreds: params.ignoreEC2RoleCreds,
				IgnoreEnvCreds:     params.ignoreEnvCreds,
				JSONOutput:         params.jsonOutput,
			}

			// Run the task
			if err := t.Run(ctx); err != nil {
				log.For(ctx).Fatal("unable to execute task", zap.Error(err))
			}
		},
	}

	// Flags
	cmd.Flags().StringVar(&params.containerPath, "in", "", "Input container path  ('-' for stdin or filename)")
	cmd.Flags().StringVar(&params.endpoint, "endpoint", "", "Set S3 compatible endpoint")
	cmd.Flags().StringVar(&params.region, "region", "", "Set service region")
	cmd.Flags().StringVar(&params.profile, "profile", "", "Set AWS profile")
	cmd.Flags().StringVar(&params.bucketName, "bucket-name", "", "AWS S3 Bucket name")
	log.CheckErr("unable to mark 'bucket-name' flag as required.", cmd.MarkFlagRequired("bucket-name"))
	cmd.Flags().StringVar(&params.objectKey, "object-key", "", "AWS S3 Bucket Object key (use container filename by default)")
	cmd.Flags().StringVar(&params.accessKeyID, "access-key-id", "", "AccessKeyID credentials")
	cmd.Flags().StringVar(&params.secretAccessKey, "secret-access-key", "", "SecretAccessKey credentials")
	cmd.Flags().StringVar(&params.sessionToken, "session-token", "", "SessionToken credentials")
	cmd.Flags().BoolVar(&params.disableSSL, "disable-ssl", false, "Disable SSL transport")
	cmd.Flags().BoolVar(&params.s3ForcePathStyle, "s3-force-path-style", false, "Disable SSL transport")
	cmd.Flags().BoolVar(&params.ignoreConfigCreds, "ignore-config-creds", false, "Disable SharedConfig credentials provider")
	cmd.Flags().BoolVar(&params.ignoreEnvCreds, "ignore-env-creds", false, "Disable Environment credentials provider")
	cmd.Flags().BoolVar(&params.ignoreEC2RoleCreds, "ignore-ec2role-creds", false, "Disable EC2Role credentials provider")
	cmd.Flags().BoolVar(&params.jsonOutput, "json", false, "Display container key as json")
	return cmd
}
