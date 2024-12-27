// Taken from https://github.com/containers/image
// Modifications Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.

package s3

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/containers/image/v5/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sha256digestHex = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	sha256digest    = "@sha256:" + sha256digestHex
	tarFixture      = "fixtures/almostempty.tar"
)

func TestTransportName(t *testing.T) {
	assert.Equal(t, "s3", Transport.Name())
}

func TestTransportParseReference(t *testing.T) {
	testParseReference(t, Transport.ParseReference)
}

func TestTransportValidatePolicyConfigurationScope(t *testing.T) {
	for _, scope := range []string{ // A semi-representative assortment of values; everything is rejected.
		"docker.io/library/busybox:notlatest",
		"docker.io/library/busybox",
		"docker.io/library",
		"docker.io",
		"",
	} {
		err := Transport.ValidatePolicyConfigurationScope(scope)
		assert.Error(t, err, scope)
	}
}

func TestParseReference(t *testing.T) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	require.NoError(t, err)
	testParseReference(t, func(refString string) (types.ImageReference, error) {
		return ParseReference(refString, cfg)
	})
}

// testParseReference is a test shared for Transport.ParseReference and ParseReference.
func testParseReference(t *testing.T, fn func(string) (types.ImageReference, error)) {
	ctx := context.Background()

	for _, c := range []struct {
		input string
		want  string
	}{
		{"s3://bucket/key", "//bucket/key"},
		{"s3://bucket/key:tag", "//bucket/key:tag"},
		{"s3://bucket/key@sha256:" + sha256digestHex, "//bucket/key@sha256:" + sha256digestHex},
		{"s3://bucket/key:tag@sha256:" + sha256digestHex, "//bucket/key:tag@sha256:" + sha256digestHex},
	} {
		ref, err := fn(c.input)
		require.NoError(t, err, c.input)
		s3ref, ok := ref.(*s3ArchiveReference)
		require.True(t, ok)
		assert.Equal(t, c.want, s3ref.StringWithinTransport())

		img, err := ref.NewImage(ctx, &types.SystemContext{})
		assert.Error(t, err)
		assert.Nil(t, img)

		s, err := ref.NewImageSource(ctx, &types.SystemContext{})
		assert.Error(t, err)
		assert.Nil(t, s)

		d, err := ref.NewImageDestination(ctx, &types.SystemContext{})
		assert.Error(t, err)
		assert.Nil(t, d)

		assert.Equal(t, "", ref.PolicyConfigurationIdentity())
		assert.Equal(t, []string{}, ref.PolicyConfigurationNamespaces())
		assert.Error(t, ref.DeleteImage(ctx, &types.SystemContext{}))
	}
}

func TestReferenceTransport(t *testing.T) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	require.NoError(t, err)
	ref, err := ParseReference("//bucket/archive.tar:nginx:latest", cfg)
	require.NoError(t, err)
	assert.Equal(t, Transport, ref.Transport())
}
