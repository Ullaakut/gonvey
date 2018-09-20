package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ullaakut/gonvey/logger"
	"github.com/rs/zerolog"
	"gopkg.in/tylerb/graceful.v1"
)

const serverAddr = "http://0.0.0.0:4242"

func main() {
	log := logger.NewZeroLog(os.Stderr)
	log.Info().Msg("gonvey is starting up")

	// TODO: Add log level to configuration
	zerolog.SetGlobalLevel(logger.ParseLevel("DEBUG"))

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
			Addr: ":8888",
			Handler: &ReverseProxy{
				log: log,
				p:   proxy,
			},
		},
	}

	// Start server
	go func() {
		log.Fatal().Err(server.ListenAndServe()).Msg("server stopped")
	}()

	log.Info().Msg("gonvey is up")

	// Wait for server to be stopped
	<-sig
	signal.Stop(sig)
	close(sig)

	log.Info().Msg("gonvey is shutting down")

	server.Stop(10 * time.Second)

	log.Info().Msg("gonvey shutdown complete")

	os.Exit(0)
}
