// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

const defaultCookieName = "beacon_is_great"

var (
	ErrMissingDatabasePath = errors.New("please set the database path")
	ErrMissingJWTSecret    = errors.New("the JWT Secret is empty")
	ErrInvalidCookieName   = errors.New("the configured cookie name is invalid")
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
	Secret     string `json:"secret"`
	CookieName string `json:"cookieName"`
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

	if cfg.JWT.CookieName == "" {
		cfg.JWT.CookieName = defaultCookieName
	}

	if err := validateCookieName(cfg.JWT.CookieName); err != nil {
		return Config{}, fmt.Errorf("error validating the cookie name: %w", err)
	}

	return cfg, nil
}

func validateCookieName(name string) error {
	pattern := regexp.MustCompile(`^(?:[A-Za-z0-9]|\+|\-|\.|\_)+$`)

	if !pattern.MatchString(name) {
		return ErrInvalidCookieName
	}

	return nil
}
