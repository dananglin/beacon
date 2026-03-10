// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package discovery

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"slices"
	"strings"
	"time"
)

var (
	ErrMismatchedClientID = errors.New("the client ID in the authorization request does not match the client ID in the client ID's metadata")
	ErrInvalidClientURL   = errors.New("the client URL in the metadata is not a prefix of the client ID")
	ErrInvalidRedirectURI = errors.New("the redirect URI in the authorization request is invalid")
)

// ValidateClientMetadata validates the received client metadata against the criteria specified in the
// IndieAuth specification.
func ValidateClientMetadata(metadata ClientIDMetadata, requestedClientID, requestedRedirectURI string) error {
	// The Client ID in the metadata must match the Client ID in the authorization request.
	if metadata.ClientID != requestedClientID {
		return ErrMismatchedClientID
	}

	// The Client URL in the metadata must be a prefix of the Client ID in the metadata.
	if !strings.HasPrefix(metadata.ClientID, metadata.ClientURI) {
		return ErrInvalidClientURL
	}

	// Parse the client ID and the redirect URI in the authorization request for further inspection.
	parsedClientID, err := url.Parse(metadata.ClientID)
	if err != nil {
		return fmt.Errorf("error parsing the client ID: %w", err)
	}

	parsedRequestedRedirectURI, err := url.Parse(requestedRedirectURI)
	if err != nil {
		return fmt.Errorf("error parsing the requested redirect URI: %w", err)
	}

	// The scheme, host and port of the requested redirect URI should match that of the client ID.
	// Otherwise the redirect URI MUST match one of the redirect URIs in the client ID metadata.
	if parsedClientID.Scheme != parsedRequestedRedirectURI.Scheme || parsedClientID.Hostname() != parsedRequestedRedirectURI.Hostname() || parsedClientID.Port() != parsedRequestedRedirectURI.Port() {
		if !slices.Contains(metadata.RedirectURIs, parsedRequestedRedirectURI.String()) {
			return ErrInvalidRedirectURI
		}
	}

	return nil
}

var (
	ErrClientIDLoopBackAddr  = errors.New("the hostname of client ID resolves to a loopback address")
	ErrClientIDInvalidScheme = errors.New("the client ID contains a non-http-based URI scheme")
)

// ValidateClientID inspects the client's ID to ensure that it doesn't resolve to
// an IP address in the internal network.
func ValidateClientID(clientID string) error {
	// Parse the client ID for inspection.
	parsedClientID, err := url.Parse(clientID)
	if err != nil {
		return fmt.Errorf("error parsing the client ID: %w", err)
	}

	if parsedClientID.Scheme != "http" && parsedClientID.Scheme != "https" {
		return ErrClientIDInvalidScheme
	}

	var addrs []net.IP

	addr := net.ParseIP(parsedClientID.Hostname())
	if addr != nil {
		addrs = []net.IP{addr}
	} else {
		// Resolve the client ID's hostname
		resolver := net.Resolver{}
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		addrs, err = resolver.LookupIP(ctx, "ip", parsedClientID.Hostname())
		if err != nil {
			return fmt.Errorf("error resolving the client ID's hostname: %w", err)
		}

	}

	for _, addr := range addrs {
		// The IP address must not resolve to a loopback address.
		if addr.IsLoopback() {
			return ErrClientIDLoopBackAddr
		}
	}

	return nil
}
