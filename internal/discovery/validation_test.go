// SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package discovery_test

import (
	"errors"
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/discovery"
)

func TestValidateClientMetadata(t *testing.T) {
	testCases := []struct {
		name        string
		clientID    string
		redirectURI string
		metadata    discovery.ClientIDMetadata
	}{
		{
			name:        "Test Case 1",
			clientID:    "https://test.website.net/.site/metadata",
			redirectURI: "https://test.website.net/redirect",
			metadata: discovery.ClientIDMetadata{
				ClientID:     "https://test.website.net/.site/metadata",
				ClientName:   "Test Website",
				ClientURI:    "https://test.website.net/",
				LogoURI:      "https://test.website.net/assets/logo.png",
				RedirectURIs: make([]string, 0),
			},
		},
		{
			name:        "Test Case 2",
			clientID:    "https://test.website.net/",
			redirectURI: "https://redirect.website.net/",
			metadata: discovery.ClientIDMetadata{
				ClientID:   "https://test.website.net/",
				ClientName: "Test Website",
				ClientURI:  "https://test.website.net/",
				LogoURI:    "https://test.website.net/assets/logo.png",
				RedirectURIs: []string{
					"https://redirect.website.net/",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, testValidClientMetadata(tc.name, tc.clientID, tc.redirectURI, tc.metadata))
	}

	errorCases := []struct {
		name        string
		clientID    string
		redirectURI string
		metadata    discovery.ClientIDMetadata
		wantError   error
	}{
		{
			name:        "Mismatched client ID",
			clientID:    "http://test.website.net/",
			redirectURI: "https://test.website.net/redirect",
			metadata: discovery.ClientIDMetadata{
				ClientID:     "https://test.website.net/.site/metadata",
				ClientName:   "Test Website",
				ClientURI:    "https://test.website.net/",
				LogoURI:      "https://test.website.net/assets/logo.png",
				RedirectURIs: make([]string, 0),
			},
			wantError: discovery.ErrMismatchedClientID,
		},
		{
			name:        "Invalid redirect URI",
			clientID:    "https://test.website.net/",
			redirectURI: "https://test2.website.net/redirect",
			metadata: discovery.ClientIDMetadata{
				ClientID:   "https://test.website.net/",
				ClientName: "Test Website",
				ClientURI:  "https://test.website.net/",
				LogoURI:    "https://test.website.net/assets/logo.png",
				RedirectURIs: []string{
					"https://redirect.website.net/",
				},
			},
			wantError: discovery.ErrInvalidRedirectURI,
		},
		{
			name:        "Invalid client URL",
			clientID:    "https://test.website.net/.site/metadata",
			redirectURI: "https://test.website.net/redirect",
			metadata: discovery.ClientIDMetadata{
				ClientID:     "https://test.website.net/.site/metadata",
				ClientName:   "Test Website",
				ClientURI:    "https://website.net/",
				LogoURI:      "https://test.website.net/assets/logo.png",
				RedirectURIs: make([]string, 0),
			},
			wantError: discovery.ErrInvalidClientURL,
		},
	}

	for _, ec := range errorCases {
		t.Run(ec.name, testInvalidClientMetadata(ec.name, ec.clientID, ec.redirectURI, ec.metadata, ec.wantError))
	}
}

func testValidClientMetadata(
	testName string,
	clientID string,
	redirectURI string,
	metadata discovery.ClientIDMetadata,
) func(t *testing.T) {
	return func(t *testing.T) {
		if err := discovery.ValidateClientMetadata(metadata, clientID, redirectURI); err != nil {
			t.Errorf(
				"FAILED test %q: Received an unexpected error validating the client's metadata: %v",
				testName,
				err,
			)
		} else {
			t.Logf("%s passed validation", testName)
		}
	}
}

func testInvalidClientMetadata(
	testName string,
	clientID string,
	redirectURI string,
	metadata discovery.ClientIDMetadata,
	wantError error,
) func(t *testing.T) {
	return func(t *testing.T) {
		if err := discovery.ValidateClientMetadata(metadata, clientID, redirectURI); err == nil {
			t.Errorf(
				"FAILED test %q: No error was received using invalid client metadata",
				testName,
			)
		} else {
			if !errors.Is(err, wantError) {
				t.Errorf(
					"FAILED test %q: Unexpected error received using invalid client metadata.\nwant: %v\ngot: %v",
					testName,
					wantError,
					err,
				)
			} else {
				t.Logf(
					"Expected error received using invalid client metadata.\ngot: %v",
					err,
				)
			}
		}
	}
}
