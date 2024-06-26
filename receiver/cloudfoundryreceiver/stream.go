// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cloudfoundryreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/cloudfoundryreceiver"

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.uber.org/zap"
)

type EnvelopeStreamFactory struct {
	rlpGatewayClient *loggregator.RLPGatewayClient
}

func newEnvelopeStreamFactory(
	ctx context.Context,
	settings component.TelemetrySettings,
	authTokenProvider *UAATokenProvider,
	httpConfig confighttp.ClientConfig,
	host component.Host) (*EnvelopeStreamFactory, error) {

	httpClient, err := httpConfig.ToClient(ctx, host, settings)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP client for Cloud Foundry RLP Gateway: %w", err)
	}

	gatewayClient := loggregator.NewRLPGatewayClient(httpConfig.Endpoint,
		loggregator.WithRLPGatewayClientLogger(zap.NewStdLog(settings.Logger)),
		loggregator.WithRLPGatewayHTTPClient(&authorizationProvider{
			logger:            settings.Logger,
			authTokenProvider: authTokenProvider,
			client:            httpClient,
		}),
	)

	return &EnvelopeStreamFactory{gatewayClient}, nil
}

func (rgc *EnvelopeStreamFactory) CreateStream(
	ctx context.Context,
	shardID string,
	telemetryType telemetryType) (loggregator.EnvelopeStream, error) {

	newShardID := shardID
	selectors := []*loggregator_v2.Selector{}
	switch telemetryType {
	case telemetryTypeLogs:
		newShardID = shardID + "_logs"
		selectors = append(selectors, &loggregator_v2.Selector{
			Message: &loggregator_v2.Selector_Log{
				Log: &loggregator_v2.LogSelector{},
			},
		})
	case telemetryTypeMetrics:
		newShardID = shardID + "_metrics"
		selectors = append(selectors, &loggregator_v2.Selector{
			Message: &loggregator_v2.Selector_Counter{
				Counter: &loggregator_v2.CounterSelector{},
			},
		})
		selectors = append(selectors, &loggregator_v2.Selector{
			Message: &loggregator_v2.Selector_Gauge{
				Gauge: &loggregator_v2.GaugeSelector{},
			},
		})
	case telemetryTypeTraces:
		newShardID = shardID + "_traces"
		//TODO
		//selectors = append(selectors, &loggregator_v2.Selector{
		//Message: &loggregator_v2.Selector_Span{
		//Span: &loggregator_v2.SpanSelector{},
		//},
		//})

	}
	stream := rgc.rlpGatewayClient.Stream(ctx, &loggregator_v2.EgressBatchRequest{
		ShardId:   newShardID,
		Selectors: selectors,
	})

	return stream, nil
}

type authorizationProvider struct {
	logger            *zap.Logger
	authTokenProvider *UAATokenProvider
	client            *http.Client
}

func (ap *authorizationProvider) Do(request *http.Request) (*http.Response, error) {
	token, err := ap.authTokenProvider.ProvideToken()
	if err == nil {
		request.Header.Set("Authorization", token)
	} else {
		ap.logger.Error("fetching authentication token", zap.Error(err))
		return nil, errors.New("obtaining authentication token for the request")
	}

	return ap.client.Do(request)
}
