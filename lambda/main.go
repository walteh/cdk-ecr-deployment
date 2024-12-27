// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	_ "cdk-ecr-deployment-handler/s3" // Install s3 transport plugin
)

const EnvLogLevel = "LOG_LEVEL"

func init() {
	s, exists := os.LookupEnv(EnvLogLevel)
	if !exists {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		lvl, err := logrus.ParseLevel(s)
		if err != nil {
			logrus.Errorf("error parsing %s: %v", EnvLogLevel, err)
		}
		logrus.SetLevel(lvl)
	}
}

func handler(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	physicalResourceID = event.PhysicalResourceID
	data = make(map[string]interface{})

	log.Printf("Event: %s", Dumps(event))

	// Load AWS config once at the start
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return physicalResourceID, data, fmt.Errorf("failed to load AWS config: %v", err)
	}

	log.Printf("AWS_ENDPOINT_URL: %s", os.Getenv("AWS_ENDPOINT_URL"))

	if os.Getenv("AWS_ENDPOINT_URL") != "" {
		cfg.BaseEndpoint = aws.String(os.Getenv("AWS_ENDPOINT_URL"))
		cfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: os.Getenv("AWS_ENDPOINT_URL"),
			}, nil
		})
	}

	if event.RequestType == cfn.RequestDelete {
		return physicalResourceID, data, nil
	}
	if event.RequestType == cfn.RequestCreate || event.RequestType == cfn.RequestUpdate {
		srcImage, err := getStrProps(event.ResourceProperties, SRC_IMAGE)
		if err != nil {
			return physicalResourceID, data, err
		}
		destImage, err := getStrProps(event.ResourceProperties, DEST_IMAGE)
		if err != nil {
			return physicalResourceID, data, err
		}
		imageArch, err := getStrPropsDefault(event.ResourceProperties, IMAGE_ARCH, "")
		if err != nil {
			return physicalResourceID, data, err
		}
		srcCreds, err := getStrPropsDefault(event.ResourceProperties, SRC_CREDS, "")
		if err != nil {
			return physicalResourceID, data, err
		}
		destCreds, err := getStrPropsDefault(event.ResourceProperties, DEST_CREDS, "")
		if err != nil {
			return physicalResourceID, data, err
		}

		srcCreds, err = parseCreds(ctx, cfg, srcCreds)
		if err != nil {
			return physicalResourceID, data, err
		}
		destCreds, err = parseCreds(ctx, cfg, destCreds)
		if err != nil {
			return physicalResourceID, data, err
		}

		log.Printf("SrcImage: %v DestImage: %v ImageArch: %v", srcImage, destImage, imageArch)

		srcRef, err := alltransports.ParseImageName(srcImage)
		if err != nil {
			return physicalResourceID, data, err
		}
		destRef, err := alltransports.ParseImageName(destImage)
		if err != nil {
			return physicalResourceID, data, err
		}

		srcOpts := NewImageOpts(srcImage, imageArch)
		srcOpts.SetCreds(srcCreds)
		srcCtx, err := srcOpts.NewSystemContext(ctx)
		if err != nil {
			return physicalResourceID, data, err
		}
		destOpts := NewImageOpts(destImage, imageArch)
		destOpts.SetCreds(destCreds)
		destCtx, err := destOpts.NewSystemContext(ctx)
		if err != nil {
			return physicalResourceID, data, err
		}

		policyContext, err := newPolicyContext()
		if err != nil {
			return physicalResourceID, data, err
		}
		defer policyContext.Destroy()

		_, err = copy.Image(ctx, policyContext, destRef, srcRef, &copy.Options{
			ReportWriter:   os.Stdout,
			DestinationCtx: destCtx,
			SourceCtx:      srcCtx,
		})
		if err != nil {
			return physicalResourceID, data, fmt.Errorf("copy image failed: %s", err.Error())
		}
	}

	return physicalResourceID, data, nil
}

func main() {
	lambda.Start(cfn.LambdaWrap(handler))
}

func newPolicyContext() (*signature.PolicyContext, error) {
	policy := &signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}}
	return signature.NewPolicyContext(policy)
}

func getStrProps(m map[string]interface{}, k string) (string, error) {
	v := m[k]
	val, ok := v.(string)
	if ok {
		return val, nil
	}
	return "", fmt.Errorf("can't get %v", k)
}

func getStrPropsDefault(m map[string]interface{}, k string, d string) (string, error) {
	v := m[k]
	if v == nil {
		return d, nil
	}
	val, ok := v.(string)
	if ok {
		return val, nil
	}
	return "", fmt.Errorf("can't get %v", k)
}

func parseCreds(ctx context.Context, cfg aws.Config, creds string) (string, error) {
	credsType := GetCredsType(creds)
	if creds == "" {
		return "", nil
	} else if (credsType == SECRET_ARN) || (credsType == SECRET_NAME) {
		secret, err := GetSecret(ctx, cfg, creds)
		return secret, err
	} else if credsType == SECRET_TEXT {
		return creds, nil
	}
	return "", fmt.Errorf("unknown creds type")
}
