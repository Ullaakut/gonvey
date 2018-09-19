package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/tylerb/graceful.v1"
)

// Proxy is a wrapper around httputil.ReverseProxy
type Proxy struct {
	p *httputil.ReverseProxy
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Gonvey", "Gonvey")
	log.Println("Forwarding request to", serverAddr, "on", r.URL, "from", r.RemoteAddr)
	p.p.ServeHTTP(w, r)
}

const serverAddr = "http://0.0.0.0:4242"

func main() {
	log.Println("gonvey starting up...")

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	endpoint, err := url.Parse(serverAddr)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(endpoint)

	server := &graceful.Server{
		NoSignalHandling: true,
		Timeout:          10 * time.Second,
		Server: &http.Server{
			Addr:    ":8888",
			Handler: &Proxy{proxy},
		},
	}

	// Start server
	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	log.Println("gonvey is up")

	// Wait for server to be stopped
	<-sig
	signal.Stop(sig)
	close(sig)

	log.Println("gonvey is shutting down")

	server.Stop(10 * time.Second)

	log.Println("gonvey shutdown complete")

	os.Exit(0)
}
