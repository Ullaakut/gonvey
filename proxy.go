package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// loadBalance just randomly gets a proxy out of the slice of proxies
// Not the smartest load balancing, but certainly the simplest
// We assume that the list of proxies is never empty, as it would have
// triggered a validation error in `config.go`
func loadBalance(proxies []*httputil.ReverseProxy) *httputil.ReverseProxy {
	// set random seed
	rand.Seed(time.Now().UnixNano())

	return proxies[rand.Intn(len(proxies))]
}

// splitPath splits the requestURI between the path and the subpath
// the path is the part of the URI that is used to match with a set of endpoints
// the subpath is the other part of the URI that will be added to the endpoint
// eg: an endpoint is bound to `/bloggo`, and a request for `/bloggo/posts` comes up, the
// request will be forwarded to the endpoint with `/posts` as its request URI
func splitPath(requestURI string, proxyMap map[string][]*httputil.ReverseProxy) (string, string, error) {
	for endpoint := range proxyMap {
		if strings.HasPrefix(requestURI, endpoint) {
			return endpoint, strings.Replace(requestURI, endpoint, "", 1), nil
		}
	}
	return "", "", fmt.Errorf("path %s is not bound to any endpoints", requestURI)
}

// MultiHostReverseProxy is a wrapper around httputil.ReverseProxy
type MultiHostReverseProxy struct {
	log *zerolog.Logger
	p   map[string][]*httputil.ReverseProxy
}

// NewMultiHostReverseProxy generates a new reverse proxy for multiple hosts from a proxy map
func NewMultiHostReverseProxy(logger *zerolog.Logger, proxyMap map[string][]string) (*MultiHostReverseProxy, error) {
	proxy := &MultiHostReverseProxy{
		p:   make(map[string][]*httputil.ReverseProxy),
		log: logger,
	}

	// create singleHostReverseProxies for each entry in the proxy map
	for path, endpoints := range proxyMap {
		for _, endpoint := range endpoints {
			url, err := url.Parse(endpoint)
			if err != nil {
				return nil, err
			}

			singleHostReverseProxy := httputil.NewSingleHostReverseProxy(url)

			// disables logging on stderr if server is unreachable
			singleHostReverseProxy.ErrorLog = log.New(ioutil.Discard, "", 0)

			proxy.p[path] = append(proxy.p[path], singleHostReverseProxy)
		}
	}

	return proxy, nil
}

func (mhrp *MultiHostReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Gonvey", "Gonvey")

	// Prometheus handles the /metrics endpoint
	if r.RequestURI == "/metrics" {
		handle := promhttp.Handler()
		handle.ServeHTTP(w, r)
		return
	}

	// get path and subpath from request URI to match with endpoint
	path, subpath, err := splitPath(r.RequestURI, mhrp.p)
	if err != nil {
		mhrp.log.Error().Str("request_uri", r.RequestURI).Msg("unknown path")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Rewrite the URI to be forwarded by removing the path used for routing
	// eg: `/bloggo/posts` becomes `/posts` redirected to an endpoint
	// registered to the `/bloggo` route
	r.URL.Path = subpath

	// Pick a random endpoint for this path
	proxy := loadBalance(mhrp.p[path])
	// Use a Gonveyor as the Transport, in order to add metrics and logging
	proxy.Transport = &Gonveyor{
		log: mhrp.log,
	}
	proxy.ServeHTTP(w, r)
}
