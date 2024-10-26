package config_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/config"
)

func TestConfig(t *testing.T) {
	want := config.Config{
		BindAddress: "127.0.0.1",
		Port:        443,
		Domain:      "auth.example.net",
		Database: config.Database{
			Path: "/app/data/indieauth.db",
		},
		JWT: config.JWT{
			Secret: "N4N6Zpwq6tCHR3CcvHmnUynQhU6R6dk0wfi3kFV1o9I0OV6l53xRxQlvQA76aYgP",
		},
	}

	configPath := filepath.Join("testdata", t.Name()+".golden")

	got, err := config.NewConfig(configPath)
	if err != nil {
		t.Fatalf(
			"FAILED test %s: Error received while loading the config: %v",
			t.Name(),
			err,
		)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf(
			"FAILED test %s: Unexpected config loaded from file, want:\n%+v\ngot:\n%+v",
			t.Name(),
			want,
			got,
		)
	} else {
		t.Logf(
			"PASSED test %s: Expected config loaded from file, got:\n%+v",
			t.Name(),
			got,
		)
	}
}
