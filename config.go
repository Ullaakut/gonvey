package main

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	v "gopkg.in/go-playground/validator.v9"
)

// Config represents the Bloggo configuration
type Config struct {
	LogLevel   string              `json:"log_level" validate:"required,eq=DEBUG|eq=INFO|eq=WARNING|eq=ERROR|eq=FATAL"`
	ServerPort uint                `json:"server_port" validate:"required,min=1,max=65535"`
	ProxyMap   map[string][]string `json:"proxy_map"`
}

// Set default values for configuration parameters
func init() {
	viper.SetDefault("log_level", "DEBUG")
	viper.SetDefault("server_port", 8888)
	viper.SetDefault("proxy_map", map[string][]string{
		"/bloggo": []string{"http://app1:4242"},
		"/test":   []string{"http://app2:4243", "http://app3:4244", "http://app4:4245"},
	})
}

// GetConfig sets the default values for the configuration and gets it from the environment/command line
func GetConfig() (Config, error) {
	var config Config

	// Override default with environment variables
	viper.SetEnvPrefix("GONVEY")
	viper.AutomaticEnv()
	viper.Unmarshal(&config)

	config.LogLevel = viper.GetString("log_level")
	config.ServerPort = uint(viper.GetInt("server_port"))
	config.ProxyMap = viper.GetStringMapStringSlice("proxy_map")

	validate := v.New()
	err := validate.Struct(config)
	if err != nil {
		return config, err
	}

	return config, nil
}

// Print prints the current configuration
func (c Config) Print(log *zerolog.Logger) {
	log.Debug().
		Str("log_level", c.LogLevel).
		Str("endpoint", fmt.Sprintf("%+v", c.ProxyMap)).
		Uint("server_port", c.ServerPort).
		Msg("configuration")
}
