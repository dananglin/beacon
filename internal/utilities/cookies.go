// SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package utilities

import "regexp"

type UnsupportedCookieNameError struct {
	name string
}

func (e UnsupportedCookieNameError) Error() string {
	return "unsupported cookie name: " + e.name
}

func ValidateCookieName(name string) error {
	pattern := regexp.MustCompile(`^(?:[A-Za-z0-9]|\+|\-|\.|\_)+$`)

	if !pattern.MatchString(name) {
		return UnsupportedCookieNameError{name: name}
	}

	return nil
}
