// SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package auth_test

import (
	"errors"
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/auth"
)

func TestVerifyAuthorizationCode(t *testing.T) {
	// The code challenge was generated manually using the following one-liner:
	// echo -n "<code>" | openssl dgst -binary -sha256 | basenc --base64url -w 0 | tr -d [===]
	codeChallenge := "ZBqCKcL6v6M9E0Fg7dddisV41DoWKs86gABnRmxVZq4"
	validCodeVerifier := "xrgtxKJG5iuK7EcS"
	t.Run("Test Valid Code Verifier", testValidCodeVerifier(codeChallenge, validCodeVerifier))

	invalidCodeVerifier := "xrgtxKJG4iuK7EcS"
	t.Run("Test Invalid Code Verifier", testInvalidCodeVerifier(codeChallenge, invalidCodeVerifier))

	t.Run("Test Unsupported Hash Method", testUnsupportedHashMethod("bad_hash_method", codeChallenge, validCodeVerifier))
}

func testValidCodeVerifier(codeChallenge, codeVerifier string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		if err := auth.VerifyAuthorizationCode("S256", codeChallenge, codeVerifier); err != nil {
			t.Errorf(
				"FAILED test %s: Authorization code validation failed.\ngot: %q",
				t.Name(),
				err.Error(),
			)
		} else {
			t.Log("Authorization code validation succeeded.")
		}
	}
}

func testInvalidCodeVerifier(codeChallenge, codeVerifier string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		err := auth.VerifyAuthorizationCode("SHA256", codeChallenge, codeVerifier)
		if err == nil {
			t.Errorf(
				"FAILED test %s: The invalid code verified was verified",
				t.Name(),
			)
		} else {
			var wantErr auth.UnverifiedAuthorizationCodeError
			if !errors.As(err, &wantErr) {
				t.Errorf(
					"FAILED test %s: Unexpected error received for using an invalid code verifier:\nwant: %q\ngot: %q",
					t.Name(),
					wantErr.Error(),
					err.Error(),
				)
			} else {
				t.Logf(
					"Expected error received for using an invalid code verifier:\ngot: %q",
					err.Error(),
				)
			}
		}
	}
}

func testUnsupportedHashMethod(hashMethod, codeChallenge, codeVerifier string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		err := auth.VerifyAuthorizationCode(hashMethod, codeChallenge, codeVerifier)
		if err == nil {
			t.Errorf(
				"FAILED test %s: The validation function did not return an error using an unsupported hash method",
				t.Name(),
			)
		} else {
			var wantErr auth.UnsupportedHashMethodError
			if !errors.As(err, &wantErr) {
				t.Errorf(
					"FAILED test %s: Unexpected error received for unsupported hashing method\nwant: %q\ngot: %q",
					t.Name(),
					wantErr.Error(),
					err.Error(),
				)
			} else {
				t.Logf(
					"Expected error received for unsupported hashing method\ngot: %q",
					err.Error(),
				)
			}
		}
	}
}
