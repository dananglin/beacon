package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrMissingDatabasePath = errors.New("please set the database path")
	ErrMissingJWTSecret    = errors.New("JWT Secret is empty")
)

type Config struct {
	BindAddress string   `json:"bindAddress"`
	Port        int32    `json:"port"`
	Domain      string   `json:"domain"`
	Database    Database `json:"database"`
	JWT         JWT      `json:"jwt"`
}

type Database struct {
	Path string `json:"path"`
}

type JWT struct {
	Secret string `json:"secret"`
}

func NewConfig(path string) (Config, error) {
	path = filepath.Clean(path)

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf(
			"unable to read the config from %q: %w",
			path,
			err,
		)
	}

	var cfg Config

	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf(
			"unable to decode the JSON data: %w",
			err,
		)
	}

	if cfg.Database.Path == "" {
		return Config{}, ErrMissingDatabasePath
	}

	if cfg.JWT.Secret == "" {
		return Config{}, ErrMissingJWTSecret
	}

	return cfg, nil
}
