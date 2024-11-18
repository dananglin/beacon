// SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func CreateBearerToken() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("unable to create random bytes: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(b)

	return token, nil
}
