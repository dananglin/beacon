// SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package auth_test

import (
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/auth"
)

func TestCreateBearerToken(t *testing.T) {
	token, err := auth.CreateBearerToken()
	if err != nil {
		t.Fatalf(
			"FAILED test %s: Received an unexpected error creating a bearer token.\ngot: %q",
			t.Name(),
			err.Error(),
		)
	} else {
		t.Logf(
			"Successfully created the bearer token.\ngot: %q",
			token,
		)
	}
}
