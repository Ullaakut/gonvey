package main

import (
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// loadBalance just randomly gets a proxy out of the slice of proxies
// Not the smartest load balancing, but certainly the simplest
func loadBalance(proxies []*httputil.ReverseProxy) *httputil.ReverseProxy {
	if len(proxies) == 0 {
		return nil
	}

	// set random seed
	rand.Seed(time.Now().UnixNano())

	return proxies[rand.Intn(len(proxies))]
}

func splitPath(requestURI string) (string, string) {
	if len(requestURI) < 1 {
		return "", ""
	}

	pathIdx := strings.Index(requestURI[1:], "/")
	path := requestURI[:pathIdx+1]

	subpath := strings.TrimLeftFunc(requestURI[1:], func(r rune) bool {
		return r != '/'
	})

	return path, subpath
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

	// TODO: This only allows paths that are `/something/`, so for example you
	// can't bind `/bloggo/api/v1/` to an endpoint and `/bloggo/api/v2` to another,
	// as the split will just match with `/bloggo` for both
	path, subpath := splitPath(r.RequestURI)

	// Rewrite the URI to be forwarded by removing the path used for routing
	// eg: `/bloggo/posts` becomes `/posts` redirected to an endpoint
	// registered to the `/bloggo` route
	r.URL.Path = subpath

	// Pick a random endpoint for this path
	proxy := loadBalance(mhrp.p[path])
	if proxy == nil {
		mhrp.log.Error().Str("path", path).Msg("unknown path")
		w.WriteHeader(http.StatusNotFound)
	} else {
		proxy.Transport = &Gonveyor{
			log: mhrp.log,
		}
		proxy.ServeHTTP(w, r)
	}
}
