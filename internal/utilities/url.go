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
	httpScheme  = "http"
	httpsScheme = "https"
)

var (
	ErrMissingHostname           = errors.New("the hostname is missing from the URL")
	ErrHostIsIPAddress           = errors.New("the hostname is an IP address")
	ErrInvalidURLScheme          = errors.New("invalid URL scheme")
	ErrURLContainsFragment       = errors.New("the URL contains a fragment")
	ErrURLContainsPort           = errors.New("the URL contains a port")
	ErrURLContainsUserInfo       = errors.New("the URL contains a username and/or a password")
	ErrURLContainsDotPathSegment = errors.New("the URL contains a single-dot or double-dot path segment")
	ErrURLHasNoPathSegment       = errors.New("the URL does not contain a path segment")
)

// ValidateAndCanonicalizeURL validates the given URL according to the indieauth
// specification. The canonicalized URL is returned after it passes the
// validation checks.
func ValidateAndCanonicalizeURL(inputURL string, allowPort bool) (string, error) {
	// This regular expression pattern is used to get the URL's scheme
	// to check if it is missing. If missing, the scheme is set to https.
	schemePattern := regexp.MustCompile(`^[a-z].*:\/\/|^[a-z].*:`)
	scheme := schemePattern.FindString(inputURL)

	if scheme == "" {
		inputURL = httpsScheme + "://" + inputURL
	}

	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", fmt.Errorf("unable to parse the URL %q: %w", inputURL, err)
	}

	if parsedURL.Path == "" {
		parsedURL.Path = "/"
	}

	if err := validateURL(parsedURL, allowPort); err != nil {
		return "", err
	}

	return parsedURL.String(), nil
}

func validateURL(inputURL *url.URL, allowPort bool) error {
	if inputURL.Scheme != httpsScheme && inputURL.Scheme != httpScheme {
		return ErrInvalidURLScheme
	}

	if inputURL.Hostname() == "" {
		return ErrMissingHostname
	}

	if inputURL.Path == "" {
		return ErrURLHasNoPathSegment
	}

	if inputURL.Fragment != "" {
		return ErrURLContainsFragment
	}

	if !allowPort && inputURL.Port() != "" {
		return ErrURLContainsPort
	}

	if ip := net.ParseIP(inputURL.Hostname()); ip != nil {
		return ErrHostIsIPAddress
	}

	if inputURL.User.String() != "" {
		return ErrURLContainsUserInfo
	}

	// This regular expression pattern is used to detect any dot paths
	// in the URLs path.
	dotPathPattern := regexp.MustCompile(`\/\.+\/`)
	if dotPathPattern.MatchString(inputURL.Path) {
		return ErrURLContainsDotPathSegment
	}

	return nil
}
