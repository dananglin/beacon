// SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package utilities_test

import (
	"errors"
	"fmt"
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
)

func TestValidateCookieName(t *testing.T) {
	t.Parallel()

	validCookieNames := []string{
		"my_custom_cookie_name",
		"my.custom_cookie.name",
		"My+Custom+Cookie+Name",
		"My-Custom.Cookie-Name",
		"My.Custom-H869e-Cookie.Name",
	}

	for ind := range validCookieNames {
		t.Run(
			fmt.Sprintf("Valid Cookie Name %d", ind+1),
			testValidCookieName(validCookieNames[ind]),
		)
	}

	invalidCookieNames := []string{
		"Bad,Cookie,Name",
		"Bad=Cookie*Name",
		`Bad""Cookie[Name]`,
		"(Bad Cookie Name)",
	}

	for ind := range invalidCookieNames {
		t.Run(
			fmt.Sprintf("Invalid Cookie Name %d", ind+1),
			testInvalidCookieName(invalidCookieNames[ind]),
		)
	}
}

func testValidCookieName(name string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		if err := utilities.ValidateCookieName(name); err != nil {
			t.Errorf(
				"FAILED test %s: Received an error validating the cookie name %q.\ngot: %q",
				t.Name(),
				name,
				err.Error(),
			)
		}
	}
}

func testInvalidCookieName(name string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		err := utilities.ValidateCookieName(name)
		if err == nil {
			t.Fatalf(
				"FAILED test %s: Failed to receive an error for unsupported cookie name %q",
				t.Name(),
				name,
			)

			return
		}

		var wantErr utilities.UnsupportedCookieNameError

		if !errors.As(err, &wantErr) {
			t.Fatalf(
				"FAILED test %s: Received an unexpected error for unsupported cookie name.\nwant: %q\ngot: %q",
				t.Name(),
				wantErr.Error(),
				err.Error(),
			)
		}
	}
}
