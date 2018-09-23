package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/tylerb/graceful.v1"
)

func main() {
	log := NewZeroLog(os.Stderr)
	log.Info().Msg("gonvey is starting up")

	config, err := GetConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("invalid configuration")
		os.Exit(1)
	}
	config.Print(log)

	zerolog.SetGlobalLevel(ParseLevel(config.LogLevel))

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	proxy, err := NewMultiHostReverseProxy(log, config.ProxyMap)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid proxy map")
		os.Exit(1)
	}

	server := &graceful.Server{
		NoSignalHandling: true,
		Timeout:          10 * time.Second,
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%d", config.ServerPort),
			Handler: proxy,
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
