package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	MySQL   MySQLConfig
	Redis   RedisConfig
	Server  ServerConfig
	Tracing TracingConfig
	// Prometheus PrometheusConfig
}

type MySQLConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type ServerConfig struct {
	Port string
}

type TracingConfig struct {
	JaegerEndpoint string
}

// type PrometheusConfig struct {
// 	MetricsEndpoint string
// 	Port            int
// }

// LoadConfig reads configuration from file or environment variables
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Read the config file
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
		return nil, err
	}

	// Unmarshal config into the Config struct
	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Failed to unmarshal configuration: %v", err)
		return nil, err
	}

	return &config, nil
}
