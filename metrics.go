package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/rs/zerolog"
)

var (
	httpRequestsResponseTime prometheus.Histogram
	httpRequestsCount        *prometheus.CounterVec
	httpRemoteAddresses      *prometheus.CounterVec
	httpResponseCodes        *prometheus.CounterVec
)

// Register the prometheus metrics/collectors
func init() {
	httpRequestsResponseTime = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "http",
		Name:      "response_time_seconds",
		Help:      "Request response times",
	})

	httpRequestsCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "http",
		Name:      "requests_count",
		Help:      "Request counter",
	}, []string{"http_method", "http_request_uri", "endpoint"})

	httpRemoteAddresses = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "http",
		Name:      "remote_addr",
		Help:      "Remote addresses counter",
	}, []string{"http_remote_addr"})

	httpResponseCodes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "http",
		Name:      "response_code",
		Help:      "Response codes counter",
	}, []string{"http_response_code"})

	prometheus.MustRegister(httpRequestsResponseTime)
	prometheus.MustRegister(httpRequestsCount)
	prometheus.MustRegister(httpRemoteAddresses)
	prometheus.MustRegister(httpResponseCodes)
}

// Send metrics to Prometheus
func pushMetrics(log *zerolog.Logger, start time.Time, request *http.Request, response *http.Response) {
	httpRequestsResponseTime.Observe(float64(time.Since(start).Nanoseconds()))

	httpRequestsCount.With(prometheus.Labels{
		"http_method":      request.Method,
		"http_request_uri": request.RequestURI,
		"endpoint":         fmt.Sprint(request.URL),
	}).Add(1)

	err := push.Collectors(
		"http", push.HostnameGroupingKey(),
		"http://metrics-gateway:9091",
		httpRequestsCount,
	)
	if err != nil {
		log.Error().Err(err).Msg("could not push response time to prometheus")
	}

	httpRemoteAddresses.With(prometheus.Labels{
		"http_remote_addr": request.RemoteAddr,
	}).Add(1)

	err = push.Collectors(
		"http", push.HostnameGroupingKey(),
		"http://metrics-gateway:9091",
		httpRemoteAddresses,
	)
	if err != nil {
		log.Error().Err(err).Msg("could not push remote addresses to prometheus")
	}

	httpResponseCodes.With(prometheus.Labels{
		"http_response_code": fmt.Sprint(response.StatusCode),
	}).Add(1)

	err = push.Collectors(
		"http", push.HostnameGroupingKey(),
		"http://metrics-gateway:9091",
		httpResponseCodes,
	)
	if err != nil {
		log.Error().Err(err).Msg("could not push response codes to prometheus")
	}
}
