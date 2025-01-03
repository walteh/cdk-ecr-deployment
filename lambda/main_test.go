// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/stretchr/testify/assert"

	_ "cdk-ecr-deployment-handler/s3"
)

func TestMain(t *testing.T) {
	t.Skip()

	ctx := context.Background()

	// reference format: s3://bucket/key[:docker-reference]
	// valid examples:
	// s3://bucket/key:nginx:latest
	// s3://bucket/key:@0

	srcImage := "s3://cdk-ecr-deployment/nginx.tar:nginx:latest"
	destImage := "dir:/tmp/nginx.dir"

	log.Printf("SrcImage: %v DestImage: %v", srcImage, destImage)

	srcRef, err := alltransports.ParseImageName(srcImage)
	assert.NoError(t, err)
	destRef, err := alltransports.ParseImageName(destImage)
	assert.NoError(t, err)

	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err)

	srcOpts := NewImageOpts(srcImage, "")
	srcCtx, err := srcOpts.NewSystemContext(ctx, cfg)
	assert.NoError(t, err)
	destOpts := NewImageOpts(destImage, "")
	destCtx, err := destOpts.NewSystemContext(ctx, cfg)
	assert.NoError(t, err)

	policyContext, err := newPolicyContext()
	assert.NoError(t, err)
	defer policyContext.Destroy()

	_, err = copy.Image(ctx, policyContext, destRef, srcRef, &copy.Options{
		ReportWriter:   os.Stdout,
		DestinationCtx: destCtx,
		SourceCtx:      srcCtx,
	})
	assert.NoError(t, err)
}

func TestNewImageOpts(t *testing.T) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err)

	srcOpts := NewImageOpts("s3://cdk-ecr-deployment/nginx.tar:nginx:latest", "arm64")
	_, err = srcOpts.NewSystemContext(ctx, cfg)
	assert.NoError(t, err)
	destOpts := NewImageOpts("dir:/tmp/nginx.dir", "arm64")
	_, err = destOpts.NewSystemContext(ctx, cfg)
	assert.NoError(t, err)
}
