package model

import (
	"encoding/json"
	"fmt"
	"os"
)

type LogLevel string

const (
	INFO  LogLevel = "INFO"
	DEBUG LogLevel = "DEBUG"
	ERROR LogLevel = "ERROR"
	WARN  LogLevel = "WARN"
)

type ENV string

const (
	DEV  ENV = "DEV"
	PROD ENV = "PROD"
)

type LogConfig struct {
	LogLevel   LogLevel `json:"log_level"`
	JsonFormat bool     `json:"json_format"`
	LogDir     string   `json:"log_dir"`
}

func DefaultLogConfig() LogConfig {
	return LogConfig{
		LogLevel:   DEBUG,
		JsonFormat: false,
		LogDir:     "",
	}
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func DefaultMongoConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     "localhost",
		Port:     27017,
		User:     "",
		Password: "",
		Name:     "gcchrcore",
	}
}

type Config struct {
	Port      int            `json:"port"`
	Env       ENV            `json:"env"`
	Pepper    string         `json:"pepper"`
	HMACKey   string         `json:"hmac_key"`
	MongoDB   DatabaseConfig `json:"mongo_db"`
	LogConfig LogConfig      `json:"log_config"`
}

func (c *Config) IsProd() bool {
	return c.Env == PROD
}

func DefaultConfig() Config {
	return Config{
		Port:      1986,
		Env:       DEV,
		Pepper:    "some-secret-random-string",
		HMACKey:   "secret-random-hmac-key",
		MongoDB:   DefaultMongoConfig(),
		LogConfig: DefaultLogConfig(),
	}
}

func LoadConfig(configReq bool) Config {
	f, err := os.Open("core.config")
	if err != nil {
		if configReq {
			panic(err)
		}
		fmt.Println("Using default config...")
		return DefaultConfig()
	}
	var c Config
	dec := json.NewDecoder(f)
	err = dec.Decode(&c)
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully loaded core.config")
	return c
}
