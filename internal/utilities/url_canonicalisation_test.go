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

func TestValidateProfileURL(t *testing.T) {
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

	for _, testCase := range slices.All(testCases) {
		t.Run(testCase.name, testValidProfileURLs(testCase.name, testCase.url, testCase.want))
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

	for _, errorCase := range slices.All(errorCases) {
		t.Run(errorCase.name, testInvalidProfileURL(errorCase.name, errorCase.url, errorCase.wantError))
	}
}

func testValidProfileURLs(testName, url, wantURL string) func(t *testing.T) {
	return func(t *testing.T) {
		canonicalisedURL, err := utilities.ValidateProfileURL(url)
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
		if _, err := utilities.ValidateProfileURL(url); err == nil {
			t.Errorf(
				"FAILED test %q: The expected error was not received using invalid profile URL %q",
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
					"PASSED test %q: Expected error received using profile URL %q: got %q",
					testName,
					url,
					err.Error(),
				)
			}
		}
	}
}
