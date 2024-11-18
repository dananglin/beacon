// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import "errors"

var (
	ErrApplicationAlreadyInitialised = errors.New("the application is already initialised")
	ErrMissingAuthorizationCode      = errors.New("invalid authorization code: the code is not present in the cache")
	ErrExpiredAuthorizationCode      = errors.New("invalid authorization code: the code has expired")
	ErrMissingGrantType              = errors.New("the required parameter 'grant_type' is missing")
	ErrInvalidProfileAccessToken     = errors.New("invalid profile access token")
	ErrInvalidFileserverPath         = errors.New("the path must not end with a '/'")
)

type MismatchedProfileIDError struct {
	authenticatedProfileID string
	profileIDInRequest     string
}

func (e MismatchedProfileIDError) Error() string {
	return "the authenticated profile ID (" +
		e.authenticatedProfileID +
		") does not match the profile ID in the authorize request (" +
		e.profileIDInRequest +
		")"
}

type MismatchedClientIDError struct {
	exchangedClientID string
	initialClientID   string
}

func (e MismatchedClientIDError) Error() string {
	return "the client ID from the exchange request (" +
		e.exchangedClientID +
		") does not match the client ID from the original authorization request (" +
		e.initialClientID +
		")"
}

type MismatchedRedirectURIError struct {
	exchangedRedirectURI string
	initialRedirectURI   string
}

func (e MismatchedRedirectURIError) Error() string {
	return "the redirect URI from the exchange request (" +
		e.exchangedRedirectURI +
		") does not match the redirect URI from the original authorization request (" +
		e.initialRedirectURI +
		")"
}

type MissingQueryValueError struct {
	parameter string
}

func (e MissingQueryValueError) Error() string {
	return "the value for '" + e.parameter + "' is missing from the URL's query"
}

type UnsupportedGrantTypeError struct {
	grantType string
}

func (e UnsupportedGrantTypeError) Error() string {
	return "unsupported grant type: " + e.grantType
}
