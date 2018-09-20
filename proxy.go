package main

import (
	"net/http"
	"net/http/httputil"

	"github.com/rs/zerolog"
)

// ReverseProxy is a wrapper around httputil.ReverseProxy
type ReverseProxy struct {
	log *zerolog.Logger
	p   *httputil.ReverseProxy
}

func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Gonvey", "Gonvey")
	p.p.Transport = &Gonveyor{
		log: p.log,
	}
	p.p.ServeHTTP(w, r)
}
