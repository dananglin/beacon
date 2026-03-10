// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package discovery_test

import (
	"errors"
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/discovery"
)

func TestValidateClientMetadata(t *testing.T) {
	validCases := []struct {
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

	for _, vc := range validCases {
		t.Run(vc.name, testValidClientMetadata(vc.name, vc.clientID, vc.redirectURI, vc.metadata))
	}

	invalidCases := []struct {
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

	for _, ic := range invalidCases {
		t.Run(ic.name, testInvalidClientMetadata(ic.name, ic.clientID, ic.redirectURI, ic.metadata, ic.wantError))
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
	wantErr error,
) func(t *testing.T) {
	return func(t *testing.T) {
		if err := discovery.ValidateClientMetadata(metadata, clientID, redirectURI); err == nil {
			t.Errorf(
				"FAILED test %q: No error was received using invalid client metadata",
				testName,
			)
		} else {
			if !errors.Is(err, wantErr) {
				t.Errorf(
					"FAILED test %q: Unexpected error received using invalid client metadata.\nwant: %v\n got: %v",
					testName,
					wantErr,
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

func TestValidateClientID(t *testing.T) {
	validCases := []struct {
		name     string
		clientID string
	}{
		{
			name:     "Test Case 1: The client ID contains an external IP address",
			clientID: "https://34.154.39.123/.site/metadata",
		},
		{
			name:     "Test Case 2: The client ID contains a hostname that resolves to external IP addresses",
			clientID: "https://example.com/id",
		},
	}

	for _, vc := range validCases {
		t.Run(vc.name, testValidClientID(vc.name, vc.clientID))
	}

	invalidCases := []struct {
		name      string
		clientID  string
		wantError error
	}{
		{
			name:      "Test Case 1: The client ID contains the IPv4 address of 127.0.0.1",
			clientID:  "http://127.0.0.1/.site/id",
			wantError: discovery.ErrClientIDLoopBackAddr,
		},
		{
			name:      "Test Case 2: The client ID contains the IPv6 address [::1]",
			clientID:  "http://[::1]/client_id",
			wantError: discovery.ErrClientIDLoopBackAddr,
		},
		{
			name:      "Test Case 3: The client ID's hostname resolves to 127.0.0.1 and/or [::1]",
			clientID:  "http://localhost:8080/.site/metadata",
			wantError: discovery.ErrClientIDLoopBackAddr,
		},
		{
			name:      "Test Case 4: The client ID contains a loopback IPv4 address",
			clientID:  "http://127.12.34.56/id",
			wantError: discovery.ErrClientIDLoopBackAddr,
		},
		{
			name:      "Test Case 5: The client ID contains a non-http-based URI scheme",
			clientID:  "ssh://example.com/id",
			wantError: discovery.ErrClientIDInvalidScheme,
		},
	}

	for _, ic := range invalidCases {
		t.Run(ic.name, testInvalidClientID(ic.name, ic.clientID, ic.wantError))
	}
}

func testValidClientID(testName, clientID string) func(t *testing.T) {
	return func(t *testing.T) {
		if err := discovery.ValidateClientID(clientID); err != nil {
			t.Errorf(
				"FAILED test %q: Received an unexpected error validating the client's ID: %v",
				testName,
				err,
			)
		} else {
			t.Logf("%s passed validation", testName)
		}
	}
}

func testInvalidClientID(testName, clientID string, wantErr error) func(t *testing.T) {
	return func(t *testing.T) {
		err := discovery.ValidateClientID(clientID)
		if err == nil {
			t.Errorf(
				"FAILED test %q: No error was received using invalid client ID",
				testName,
			)

			return
		}

		if !errors.Is(err, wantErr) {
			t.Errorf(
				"FAILED test %q: Unexpected error received using invalid client ID.\nwant: %v\n got: %v",
				testName,
				wantErr,
				err,
			)
		} else {
			t.Logf(
				"Expected error received using invalid client ID.\ngot: %v",
				err,
			)
		}
	}
}
