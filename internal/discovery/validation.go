// SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package discovery

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var (
	ErrMismatchedClientID = errors.New("the client ID in the authorization request does not match the client ID in the client ID's metadata")
	ErrInvalidClientURL   = errors.New("the client URL in the metadata is not a prefix of the client ID")
	ErrInvalidRedirectURI = errors.New("the redirect URI in the authorization request is invalid")
)

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
		validRedirectURI := false

		for ind := range metadata.RedirectURIs {
			if parsedRequestedRedirectURI.String() == metadata.RedirectURIs[ind] {
				validRedirectURI = true

				break
			}
		}

		if !validRedirectURI {
			return ErrInvalidRedirectURI
		}
	}

	return nil
}
