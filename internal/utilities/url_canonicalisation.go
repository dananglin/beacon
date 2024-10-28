// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package utilities

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
)

const (
	httpScheme  = "http://"
	httpsScheme = "https://"
)

var (
	ErrMissingHostname     = errors.New("the hostname is missing from the URL")
	ErrHostIsIPAddress     = errors.New("the hostname is an IP address")
	ErrInvalidURLScheme    = errors.New("invalid URL scheme")
	ErrURLContainsFragment = errors.New("the URL contains a fragment")
	ErrURLContainsPort     = errors.New("the URL contains a port")
	ErrURLContainsUserInfo = errors.New("the URL contains a username and/or a password")
)

// ValidateProfileURL validates the given profile URL according to the indieauth
// specification. ValidateProfileURL returns the canonicalised profile URL after
// validation checks.
func ValidateProfileURL(profileURL string) (string, error) {
	// Using regex to get and validate the scheme.
	// If its missing then set the scheme to https
	pattern := regexp.MustCompile(`^[a-z].*:\/\/|^[a-z].*:`)
	scheme := pattern.FindString(profileURL)

	if scheme == "" {
		profileURL = httpsScheme + profileURL
	} else if scheme != httpsScheme && scheme != httpScheme {
		return "", ErrInvalidURLScheme
	}

	parsedProfileURL, err := url.Parse(profileURL)
	if err != nil {
		return "", fmt.Errorf("unable to parse the URL %q: %w", profileURL, err)
	}

	if parsedProfileURL.Hostname() == "" {
		return "", ErrMissingHostname
	}

	if ip := net.ParseIP(parsedProfileURL.Hostname()); ip != nil {
		return "", ErrHostIsIPAddress
	}

	if parsedProfileURL.Fragment != "" {
		return "", ErrURLContainsFragment
	}

	if parsedProfileURL.Port() != "" {
		return "", ErrURLContainsPort
	}

	if parsedProfileURL.User.String() != "" {
		return "", ErrURLContainsUserInfo
	}

	if parsedProfileURL.Scheme == "" {
		parsedProfileURL.Scheme = "https"
	}

	if parsedProfileURL.Path == "" {
		parsedProfileURL.Path = "/"
	}

	return parsedProfileURL.String(), nil
}
