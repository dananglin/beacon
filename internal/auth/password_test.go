package auth_test

import (
	"testing"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/auth"
)

func TestHashPassword(t *testing.T) {
	testPassword := "w0TC5HCJrw66HXt1"
	testIncorrectPassword := "tBRxfM2s86cKt8JC"

	hashedPassword, err := auth.HashPassword(testPassword)
	if err != nil {
		t.Fatalf("Unable to hash the test password: %v", err)
	} else {
		t.Logf("Test password hashed successfully: got %s", hashedPassword)
	}

	if err := auth.CheckPasswordHash(hashedPassword, testPassword); err != nil {
		t.Errorf("CheckPasswordHash failed on correct password, error received: %v", err)
	} else {
		t.Log("CheckPasswordHash passed on correct password.")
	}

	if err := auth.CheckPasswordHash(hashedPassword, testIncorrectPassword); err == nil {
		t.Error("CheckPasswordHash unexpectedly passed on incorrect password.")
	} else {
		t.Log("CheckPasswordHash expectedly failed on incorrect password.")
	}
}
