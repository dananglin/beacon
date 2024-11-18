// SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

type UnsupportedHashMethodError struct {
	method string
}

func (e UnsupportedHashMethodError) Error() string {
	return "unsupported hashing method: " + e.method
}

type UnverifiedAuthorizationCodeError struct{}

func (e UnverifiedAuthorizationCodeError) Error() string {
	return "the authorization code is unverified"
}

func VerifyAuthorizationCode(hashMethod, codeChallenge, codeVerifier string) error {
	var hashedCodeVerifier [32]byte

	switch strings.ToLower(hashMethod) {
	case "sha256", "s256":
		hashedCodeVerifier = sha256Hash(codeVerifier)
	default:
		return UnsupportedHashMethodError{method: hashMethod}
	}

	encodedCodeVerifier := base64.RawURLEncoding.EncodeToString(hashedCodeVerifier[:])

	if encodedCodeVerifier != codeChallenge {
		return UnverifiedAuthorizationCodeError{}
	}

	return nil
}

func sha256Hash(code string) [32]byte {
	return sha256.Sum256([]byte(code))
}
