package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// Gonveyor is an HTTP transport layer in charge of the metrics and logging of requests
// that are going through the proxy
type Gonveyor struct {
	log *zerolog.Logger
}

// RoundTrip is the method used by ServeHTTP to handle the incoming HTTP requests
// We overload it to be able to add logging and metrics on both the request and the response
func (g *Gonveyor) RoundTrip(request *http.Request) (*http.Response, error) {
	start := time.Now()

	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		g.log.Error().
			Err(err).
			Str("http_method", request.Method).
			Str("http_remote_addr", request.RemoteAddr).
			Str("http_request_uri", request.RequestURI).
			Msg("endpoint not reachable")
		return nil, err
	}
	elapsed := time.Since(start)

	g.log.Info().
		Str("http_remote_addr", request.RemoteAddr).
		Str("http_method", request.Method).
		Str("http_request_url", fmt.Sprint(request.URL)).
		Int64("http_request_duration", elapsed.Nanoseconds()).
		Int("http_response_code", response.StatusCode).
		Msg("request proxied")

	pushMetrics(g.log, start, request, response)

	return response, nil
}
