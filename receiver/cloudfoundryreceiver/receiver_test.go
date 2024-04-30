// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cloudfoundryreceiver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"
)

// Test to make sure a new receiver can be created properly, started and shutdown with the default config
func TestDefaultValidReceiver(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	params := receivertest.NewNopCreateSettings()

	// foreach newcloudfoundrymetricsreceiver, newcloudfoundrylogsreceiver, newcloudfoundrytracesreceiver
	// do the testing
	for _, createReceiverFunc := range []func() *cloudFoundryReceiver{
		func() *cloudFoundryReceiver {
			return newCloudFoundryMetricsReceiver(params, *cfg, consumertest.NewNop())
		},
		func() *cloudFoundryReceiver {
			return newCloudFoundryLogsReceiver(params, *cfg, consumertest.NewNop())
		},
		func() *cloudFoundryReceiver {
			return newCloudFoundryTracesReceiver(params, *cfg, consumertest.NewNop())
		},
	} {
		receiver, err := createReceiverFunc()
		require.NoError(t, err)
		require.NotNil(t, receiver, "receiver creation failed")

		// Test start
		ctx := context.Background()
		err = receiver.Start(ctx, componenttest.NewNopHost())
		require.NoError(t, err)

		// Test shutdown
		err = receiver.Shutdown(ctx)
		require.NoError(t, err)

	}
}

// test template, where we pass a function for creating a receiver
func testValidReceiver(t *testing.T, createReceiverFunc func() *cloudFoundryReceiver) {
	// Test to make sure a new receiver can be created properly, started and shutdown
	receiver := createReceiverFunc()
	require.NotNil(t, receiver, "receiver creation failed")

	// Test start
	ctx := context.Background()
	err := receiver.Start(ctx, componenttest.NewNopHost())
	require.NoError(t, err)

	// Test shutdown
	err = receiver.Shutdown(ctx)
	require.NoError(t, err)
}
