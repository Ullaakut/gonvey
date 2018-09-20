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
			Str("client", request.RemoteAddr).
			Str("endpoint", fmt.Sprint(request.Host)).
			Str("path", request.RequestURI).
			Msg("endpoint not reachable")
		return nil, err
	}
	elapsed := time.Since(start)

	g.log.Info().
		Str("client", request.RemoteAddr).
		Str("request_addr", fmt.Sprint(request.URL)).
		Int64("request_duration", elapsed.Nanoseconds()).
		Int("status", response.StatusCode).
		Msg("request proxied")

	return response, err
}
