package main

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	v "gopkg.in/go-playground/validator.v9"
)

// Config represents the Bloggo configuration
type Config struct {
	LogLevel   string              `json:"log_level" validate:"required,eq=DEBUG|eq=INFO|eq=WARNING|eq=ERROR|eq=FATAL" config:"log_level,watch"`
	ServerPort uint                `json:"server_port" validate:"required,min=1,max=65535" config:"server_port"`
	ProxyMap   map[string][]string `json:"proxy_map"`

	// Contains a JSON-encoded string representing the ProxyMap. Needed for consul
	ProxyMapStr string `config:"proxy_map,watch,restart"`
}

// Set default values for configuration parameters
func init() {
	defaultProxyMap := map[string][]string{
		"/bloggo": []string{"http://0.0.0.0:4242"},
		"/test":   []string{"http://0.0.0.0:4243", "http://0.0.0.0:4244", "http://0.0.0.0:4245"},
	}

	mapStr, _ := json.Marshal(defaultProxyMap)

	viper.SetDefault("log_level", "DEBUG")
	viper.SetDefault("server_port", 8888)
	viper.SetDefault("proxy_map", defaultProxyMap)
	viper.SetDefault("proxy_map_str", string(mapStr))
	fmt.Println("Setting proxy map str to ", string(mapStr))
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
