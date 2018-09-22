package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	c "github.com/Ullaakut/gonvey/config"
	"github.com/Ullaakut/gonvey/logger"

	"github.com/rs/zerolog"
	v "gopkg.in/go-playground/validator.v9"
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

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// Setup consul KV store for live configuration
	var source *c.Consul
	startTime := time.Now()
	err = try(log, 2*time.Second, func() error {
		source, err = c.NewConsul()
		return err
	}, func() bool {
		return time.Now().Sub(startTime) < 30*time.Second
	})
	if err != nil {
		log.Fatal().Err(err).Msg("could not initialize consul client")
		os.Exit(1)
	}
	err = source.Set(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("could not populate consul client with config")
		os.Exit(1)
	}

	zerolog.SetGlobalLevel(logger.ParseLevel(config.LogLevel))

	server, err := createProxy(log, config)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid proxy map")
		os.Exit(1)
	}

	// Watch consul KV store values to detect changes in configuration
	source.Watch(&config, func(requiresRestart bool, err error) {
		if err != nil {
			log.Error().Str("error", err.Error()).Msg("config watch error")
			return
		}

		validate := v.New()
		err = validate.Struct(config)
		if err != nil {
			log.Error().Err(err).Msg("ignoring invalid configuration update in consul")
			return
		}

		zerolog.SetGlobalLevel(logger.ParseLevel(config.LogLevel))

		if requiresRestart {
			var err error
			err = json.Unmarshal([]byte(config.ProxyMapStr), &config.ProxyMap)

			fmt.Println("PROXY MAP STR", config.ProxyMapStr)
			fmt.Printf("PROXY MAP %+v\n", config.ProxyMap)

			newServer, err := createProxy(log, config)
			if err != nil {
				log.Error().Err(err).Msg("invalid proxy map in new configuration")
				os.Exit(1)
			}

			log.Warn().Msg("change in configuration requires server restart")
			server.Stop(10 * time.Second)
			server = newServer
			go log.Fatal().Err(server.ListenAndServe()).Msg("server stopped")
			log.Info().Msg("server successfully restarted with new configuration")
		}

		log.Info().Msg("configuration updated successfully")
		config.Print(log)
	})

	go log.Fatal().Err(server.ListenAndServe()).Msg("server stopped")

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

func createProxy(log *zerolog.Logger, config Config) (*graceful.Server, error) {
	proxy, err := NewMultiHostReverseProxy(log, config.ProxyMap)
	if err != nil {
		return nil, err
	}

	server := &graceful.Server{
		NoSignalHandling: true,
		Timeout:          10 * time.Second,
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%d", config.ServerPort),
			Handler: proxy,
		},
	}
	return server, nil
}

// try tries to execute a given function
// if it fails, it will keep retrying until the given shouldRetry function returns false
func try(logger *zerolog.Logger, retryDelay time.Duration, fn func() error, shouldRetry func() bool) error {
	for {
		err := fn()
		if err == nil {
			return nil
		}

		if !shouldRetry() {
			logger.Error().Err(err).Msg("operation failed too many times, aborting")
			return err
		}

		logger.Error().Err(err).Msg("operation failed, will retry")
		time.Sleep(retryDelay)
	}
}
