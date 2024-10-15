package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	BindAddress string `json:"bindAddress"`
	Port        int32  `json:"port"`
	Domain      string `json:"domain"`
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

	return cfg, nil
}
