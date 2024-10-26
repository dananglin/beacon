package auth_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

func TestJWTWithValidTokenSecret(t *testing.T) {
	var (
		testProfileID    = "https://billjones.example.net/"
		testTokenSecret  = "hBrAMcp2o4leBGFx3tGxSNIarYizdWZn"
		testTokenVersion = 1010
		expiresIn        = 10 * time.Second
	)

	signedToken, err := auth.CreateJWT(testProfileID, testTokenSecret, testTokenVersion, expiresIn)
	if err != nil {
		t.Fatalf(
			"FAILED test %s: Received an error while attempting to create the JWT token: %v",
			t.Name(),
			err,
		)
	} else {
		t.Logf(
			"Successfully created the JWT token, got: %s",
			signedToken,
		)
	}

	got, err := auth.ValidateJWT(signedToken, testTokenSecret)
	if err != nil {
		t.Fatalf(
			"FAILED test %s: Received an error attempting to validate the JWT token: %v",
			t.Name(),
			err,
		)
	}

	want := auth.ValidateJWTResults{
		TokenVersion: 1010,
		ProfileID:    "https://billjones.example.net/",
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf(
			"FAILED test %s: Unexpected results returned after validating the JWT token:\nwant: %+v\ngot: %+v",
			t.Name(),
			want,
			got,
		)
	} else {
		t.Logf(
			"Expected results returned after validating the JWT token:\ngot: %+v",
			got,
		)
	}
}

func TestJWTWithInvalidTokenSecret(t *testing.T) {
	var (
		testProfileID        = "https://billjones.example.net/"
		testTokenSecret      = "hBrAMcp2o4leBGFx3tGxSNIarYizdWZn"
		incorrectTokenSecret = "RtIZag4P6v3m5lOlPEd1u7SGXgM0uhIH"
		testTokenVersion     = 1010
		expiresIn            = 10 * time.Second
	)

	signedToken, err := auth.CreateJWT(testProfileID, testTokenSecret, testTokenVersion, expiresIn)
	if err != nil {
		t.Fatalf(
			"FAILED test %s: Received an error while attempting to create the JWT token: %v",
			t.Name(),
			err,
		)
	} else {
		t.Logf(
			"Successfully created the JWT token, got: %s",
			signedToken,
		)
	}

	if _, err := auth.ValidateJWT(signedToken, incorrectTokenSecret); err == nil {
		t.Errorf(
			"FAILED test %s: Token validation unexpectedly passed with INCORRECT token secret",
			t.Name(),
		)
	} else {
		if !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			t.Errorf(
				"FAILED test %s: Unexpected error returned after validating the token, got %q",
				t.Name(),
				err.Error(),
			)
		} else {
			t.Log("Token validation expectedly failed with incorrect token secret")
		}
	}
}

func TestJWTWithExpiredTokenSecret(t *testing.T) {
	var (
		testProfileID    = "https://billjones.example.net/"
		testTokenSecret  = "hBrAMcp2o4leBGFx3tGxSNIarYizdWZn"
		testTokenVersion = 1010
		expiresIn        = 10 * time.Millisecond
	)

	signedToken, err := auth.CreateJWT(testProfileID, testTokenSecret, testTokenVersion, expiresIn)
	if err != nil {
		t.Fatalf(
			"FAILED test %s: Received an error while attempting to create the JWT token: %v",
			t.Name(),
			err,
		)
	} else {
		t.Logf(
			"Successfully created the JWT token, got: %s",
			signedToken,
		)
	}

	time.Sleep(100 * time.Millisecond)

	if _, err := auth.ValidateJWT(signedToken, testTokenSecret); err == nil {
		t.Errorf(
			"FAILED test %s: Token validation unexpectedly passed with the EXPIRED token",
			t.Name(),
		)
	} else {
		if !errors.Is(err, jwt.ErrTokenExpired) {
			t.Errorf(
				"FAILED test %s: Unexpected error returned after validating the token, got %q",
				t.Name(),
				err.Error(),
			)
		} else {
			t.Log("Token validation expectedly failed with the expired token")
		}
	}
}
