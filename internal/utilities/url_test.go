// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package utilities_test

import (
	"errors"
	"slices"
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
)

func TestValidateAndCanonicalizeURL(t *testing.T) {
	testCases := []struct {
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

	for _, tc := range slices.All(testCases) {
		t.Run(tc.name, testValidProfileURLs(tc.name, tc.url, tc.want))
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

	for _, ec := range slices.All(errorCases) {
		t.Run(ec.name, testInvalidProfileURL(ec.name, ec.url, ec.wantError))
	}
}

func testValidProfileURLs(testName, url, wantURL string) func(t *testing.T) {
	return func(t *testing.T) {
		canonicalisedURL, err := utilities.ValidateAndCanonicalizeURL(url)
		if err != nil {
			t.Fatalf("FAILED test %q: %v", testName, err)
		}

		if canonicalisedURL != wantURL {
			t.Errorf("FAILED test %q: want %s, got %s", testName, wantURL, canonicalisedURL)
		} else {
			t.Logf("PASSED test %q: got %s", testName, canonicalisedURL)
		}
	}
}

func testInvalidProfileURL(testName, url string, wantError error) func(t *testing.T) {
	return func(t *testing.T) {
		if _, err := utilities.ValidateAndCanonicalizeURL(url); err == nil {
			t.Errorf(
				"FAILED test %q: No error was received using invalid profile URL %q",
				testName,
				url,
			)
		} else {
			if !errors.Is(err, wantError) {
				t.Errorf(
					"FAILED test %q: Unexpected error received using profile URL %q: got %q",
					testName,
					url,
					err.Error(),
				)
			} else {
				t.Logf(
					"Expected error received using profile URL %q: got %q",
					url,
					err.Error(),
				)
			}
		}
	}
}

func TestValidateClientURL(t *testing.T) {
	testCases := []struct {
		name string
		url  string
	}{
		{
			name: "Canonicalized client URL",
			url:  "https://app.example.party/",
		},
		{
			name: "Canonicalized client URL with port",
			url:  "https://app.example.party:8443/",
		},
	}

	for _, tc := range slices.All(testCases) {
		t.Run(tc.name, testValidClientURL(tc.name, tc.url))
	}
}

func testValidClientURL(testName, url string) func(t *testing.T) {
	return func(t *testing.T) {
		if err := utilities.ValidateClientURL(url); err != nil {
			t.Fatalf(
				"FAILED test %s: unexpected error received after validating %s: %v",
				testName,
				url,
				err,
			)
		} else {
			t.Logf(
				"Successfully validated the client URL: %s",
				url,
			)
		}
	}
}
