// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package config_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/config"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	testCases := []config.Config{
		{
			BindAddress:             "127.0.0.1",
			Port:                    443,
			Domain:                  "auth.example.net",
			GracefulShutdownTimeout: 10,
			Database: config.Database{
				Path: "/app/data/indieauth.db",
			},
			JWT: config.JWT{
				Secret:     "N4N6Zpwq6tCHR3CcvHmnUynQhU6R6dk0wfi3kFV1o9I0OV6l53xRxQlvQA76aYgP",
				CookieName: "my_jwt_cookie",
			},
		},
		{
			BindAddress:             "127.0.0.1",
			Port:                    443,
			Domain:                  "auth.example.net",
			GracefulShutdownTimeout: 30,
			Database: config.Database{
				Path: "/app/data/indieauth.db",
			},
			JWT: config.JWT{
				Secret:     "vrDFbzgiWEyWn21YLAo0DDVm4pO0CihJhDDZZArxKu0J8w0d-8FtKlt1tCsJFk",
				CookieName: "beacon_is_great",
			},
		},
	}

	for ind, tc := range testCases {
		path := fmt.Sprintf("testdata/%s_%d.golden", t.Name(), ind)
		t.Run(
			fmt.Sprintf("Test case %d", ind+1),
			testConfig(path, tc),
		)
	}

	errorCases := []struct {
		path    string
		wantErr error
	}{
		{
			path:    "testdata/MissingJWTSecret.golden",
			wantErr: config.ErrMissingJWTSecret,
		},
		{
			path:    "testdata/BadCookieName.golden",
			wantErr: config.ErrInvalidCookieName,
		},
		{
			path:    "testdata/MissingDatabasePath.golden",
			wantErr: config.ErrMissingDatabasePath,
		},
	}

	for ind, ec := range errorCases {
		t.Run(
			fmt.Sprintf("Error case %d", ind+1),
			testBadConfig(ec.path, ec.wantErr),
		)
	}
}

func testConfig(path string, wantConfig config.Config) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		gotConfig, err := config.NewConfig(path)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Error received while loading the config: %v",
				t.Name(),
				err,
			)
		}

		if !reflect.DeepEqual(wantConfig, gotConfig) {
			t.Errorf(
				"FAILED test %s: Unexpected config loaded from file.\nwant: %+v\ngot: %+v",
				t.Name(),
				wantConfig,
				gotConfig,
			)
		} else {
			t.Logf(
				"Expected config loaded from file.\ngot: %+v",
				gotConfig,
			)
		}
	}
}

func testBadConfig(path string, wantErr error) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		_, err := config.NewConfig(path)
		if err == nil {
			t.Fatalf(
				"FAILED test %s: Did not receive an error for invalid config.",
				t.Name(),
			)
		}

		if !errors.Is(err, wantErr) {
			t.Errorf(
				"FAILED test %s: Unexpected error received for invalid config.\nwant: %q\ngot: %q",
				t.Name(),
				wantErr.Error(),
				err.Error(),
			)
		} else {
			t.Logf(
				"Expected error received for invalid config.\ngot: %q",
				err.Error(),
			)
		}
	}
}
