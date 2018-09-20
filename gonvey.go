package main

import (
	"fmt"
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

func main() {
	log := logger.NewZeroLog(os.Stderr)
	log.Info().Msg("gonvey is starting up")

	config, err := GetConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("invalid configuration")
		os.Exit(1)
	}
	config.Print(log)

	zerolog.SetGlobalLevel(logger.ParseLevel(config.LogLevel))

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	endpoint, err := url.Parse(config.Endpoint)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid endpoint")
		os.Exit(1)
	}

	proxy := httputil.NewSingleHostReverseProxy(endpoint)

	server := &graceful.Server{
		NoSignalHandling: true,
		Timeout:          10 * time.Second,
		Server: &http.Server{
			Addr: fmt.Sprintf(":%d", config.ServerPort),
			Handler: &ReverseProxy{
				log: log,
				p:   proxy,
			},
		},
	}

	// Start server
	go func() {
		log.Fatal().Err(server.ListenAndServe()).Msg("server stopped")
		os.Exit(1)
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
