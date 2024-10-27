// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package utilities_test

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
)

func TestParseArgs(t *testing.T) {
	testCases := []struct {
		name string
		args []string
		want utilities.ParsedArgs
	}{
		{
			name: "One argument",
			args: []string{"version"},
			want: utilities.ParsedArgs{
				Name: "version",
				Args: []string{},
			},
		},
		{
			name: "Many arguments",
			args: []string{"serve", "--config", "/app/config/cfg.json"},
			want: utilities.ParsedArgs{
				Name: "serve",
				Args: []string{"--config", "/app/config/cfg.json"},
			},
		},
	}

	for _, testCase := range slices.All(testCases) {
		t.Run(testCase.name, testParseArgs(testCase.name, testCase.args, testCase.want))
	}

	errorCases := []struct {
		name      string
		args      []string
		wantError error
	}{
		{
			name:      "No arguments provided",
			args:      make([]string, 0),
			wantError: utilities.ErrNoArgumentsProvided,
		},
	}

	for _, errorCase := range slices.All(errorCases) {
		t.Run(errorCase.name, testParseArgsErrors(errorCase.name, errorCase.args, errorCase.wantError))
	}
}

func testParseArgs(testName string, args []string, want utilities.ParsedArgs) func(t *testing.T) {
	return func(t *testing.T) {
		got, err := utilities.ParseArgs(args)
		if err != nil {
			t.Fatalf("FAILED test %q: %v", testName, err)
		}

		if !reflect.DeepEqual(want, got) {
			t.Errorf(
				"FAILED test %q: want %+v, got %+v",
				testName,
				want,
				got,
			)
		} else {
			t.Logf(
				"PASSED test %q: got %+v",
				testName,
				got,
			)
		}
	}
}

func testParseArgsErrors(testName string, args []string, wantError error) func(t *testing.T) {
	return func(t *testing.T) {
		if _, err := utilities.ParseArgs(args); err == nil {
			t.Errorf(
				"FAILED test %q: The expected error was not received",
				testName,
			)
		} else if !errors.Is(err, wantError) {
			t.Errorf(
				"FAILED test %q: Unexpected error received: want %q, got %q",
				testName,
				wantError.Error(),
				err.Error(),
			)
		} else {
			t.Logf(
				"PASSED test %q: Expected error received: got %q",
				testName,
				err.Error(),
			)
		}
	}
}
