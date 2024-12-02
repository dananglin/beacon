// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package utilities_test

import (
	"errors"
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
)

func TestValidateAndCanonicalizeURL(t *testing.T) {
	testProfileIDs := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "Canonicalised URL",
			url:  "https://barry.example.org/",
			want: "https://barry.example.org/",
		},
		{
			name: "Canonicalised URL with path",
			url:  "https://example.org/username/barry",
			want: "https://example.org/username/barry",
		},
		{
			name: "Canonicalised URL with query string",
			url:  "http://example.org/users?id=1001",
			want: "http://example.org/users?id=1001",
		},
		{
			name: "Non-canonicalised URL with missing scheme",
			url:  "barry.example.org/",
			want: "https://barry.example.org/",
		},
		{
			name: "Non-canonicalised URL with missing path",
			url:  "http://barry.example.org",
			want: "http://barry.example.org/",
		},
	}

	for _, tc := range testProfileIDs {
		t.Run(tc.name, testValidURL(tc.url, tc.want, false))
	}

	testClientIDs := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "Non-canonicalised client ID with missing path",
			url:  "http://app.test.example",
			want: "http://app.test.example/",
		},
		{
			name: "Canonicalized client ID with port",
			url:  "https://app.test.example:8443",
			want: "https://app.test.example:8443/",
		},
	}

	for _, tc := range testClientIDs {
		t.Run(tc.name, testValidURL(tc.url, tc.want, true))
	}

	errorCases := []struct {
		name      string
		url       string
		wantError error
	}{
		{
			name:      "URL using the mailto scheme",
			url:       "mailto:barry@example.org",
			wantError: utilities.ErrInvalidURLScheme,
		},
		{
			name:      "URL using a non-http scheme",
			url:       "postgres://db_user:db_password@some_db_server:5432/db",
			wantError: utilities.ErrInvalidURLScheme,
		},
		{
			name:      "URL containing a port",
			url:       "http://barry.example.org:80/",
			wantError: utilities.ErrURLContainsPort,
		},
		{
			name:      "URL containing a fragment",
			url:       "https://barry.example.org/#fragment",
			wantError: utilities.ErrURLContainsFragment,
		},
		{
			name:      "URL host is an IP address",
			url:       "https://192.168.82.56/",
			wantError: utilities.ErrHostIsIPAddress,
		},
		{
			name:      "URL with a missing host",
			url:       "https:///",
			wantError: utilities.ErrMissingHostname,
		},
		{
			name:      "URL with userinfo in it",
			url:       "https://username:P@s$w0rD@example.org/",
			wantError: utilities.ErrURLContainsUserInfo,
		},
		{
			name:      "URL with a single-dot path segment",
			url:       "https://example.org/username/./barry",
			wantError: utilities.ErrURLContainsDotPathSegment,
		},
		{
			name:      "URL with a double-dot path segment",
			url:       "https://example.org/username/../barry",
			wantError: utilities.ErrURLContainsDotPathSegment,
		},
	}

	for _, ec := range errorCases {
		t.Run(ec.name, testInvalidURL(ec.url, ec.wantError))
	}
}

func testValidURL(url, wantURL string, allowPort bool) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		canonicalisedURL, err := utilities.ValidateAndCanonicalizeURL(url, allowPort)
		if err != nil {
			t.Fatalf(
				"FAILED test %s: Unexpected error received after canonicalizing URL %q.\ngot %q",
				t.Name(),
				url,
				err.Error(),
			)
		}

		if canonicalisedURL != wantURL {
			t.Errorf(
				"FAILED test %s: Unexpected canonicalized URL returned.\nwant: %q\ngot: %q",
				t.Name(),
				wantURL,
				canonicalisedURL,
			)
		} else {
			t.Logf(
				"Expected canonicalized URL returned.\ngot: %q",
				canonicalisedURL,
			)
		}
	}
}

func testInvalidURL(url string, wantErr error) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		_, err := utilities.ValidateAndCanonicalizeURL(url, false)
		if err == nil {
			t.Fatalf(
				"FAILED test %s: No error was received using invalid profile URL %q",
				t.Name(),
				url,
			)

			return
		}

		if !errors.Is(err, wantErr) {
			t.Errorf(
				"FAILED test %s: Unexpected error received using invalid profile URL %q.\nwant something like: %q\ngot: %q",
				t.Name(),
				url,
				wantErr.Error(),
				err.Error(),
			)
		} else {
			t.Logf(
				"Expected error received using invalid profile URL %q.\ngot: %q",
				url,
				err.Error(),
			)
		}
	}
}
