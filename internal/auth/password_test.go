// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package auth_test

import (
	"errors"
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/auth"
)

func TestHashPassword(t *testing.T) {
	testPassword := "w0TC5HCJrw66HXt1"
	testIncorrectPassword := "tBRxfM2s86cKt8JC"

	hashedPassword, err := auth.HashPassword(testPassword)
	if err != nil {
		t.Fatalf(
			"FAILED test %s: Unable to hash the test password: %v",
			t.Name(),
			err,
		)
	} else {
		t.Logf("Test password hashed successfully: got %s", hashedPassword)
	}

	err = auth.CheckPasswordHash(hashedPassword, testPassword)
	if err != nil {
		t.Errorf(
			"FAILED test %s: CheckPasswordHash failed on correct password, error received: %v",
			t.Name(),
			err,
		)
	} else {
		t.Log("CheckPasswordHash passed on correct password.")
	}

	err = auth.CheckPasswordHash(hashedPassword, testIncorrectPassword)
	if err == nil {
		t.Errorf(
			"FAILED test %s: CheckPasswordHash unexpectedly passed on incorrect password.",
			t.Name(),
		)
	} else {
		var wantErr auth.IncorrectPasswordError
		if !errors.As(err, &wantErr) {
			t.Errorf(
				"FAILED test %s: Unexpected error received from CheckPasswordHash for the incorrect password.\nwant: %v\n got: %v",
				t.Name(),
				wantErr,
				err,
			)
		} else {
			t.Logf(
				"Expected error received from CheckPasswordHash for the incorrect password.\ngot: %v",
				err,
			)
		}
	}
}
