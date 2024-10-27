// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package auth

import (
	"errors"
	"fmt"
	"time"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/info"
	"github.com/golang-jwt/jwt/v5"
)

type customClaims struct {
	TokenVersion int `json:"tokenVersion"`
	jwt.RegisteredClaims
}

func CreateJWT(profileID, tokenSecret string, tokenVersion int, expiresIn time.Duration) (string, error) {
	timestamp := time.Now().UTC()
	expiry := timestamp.Add(expiresIn)

	claims := customClaims{
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    info.ApplicationName,
			IssuedAt:  jwt.NewNumericDate(timestamp),
			ExpiresAt: jwt.NewNumericDate(expiry),
			Subject:   profileID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", fmt.Errorf("token signing failed: %w", err)
	}

	return signedToken, nil
}

type ValidateJWTResults struct {
	TokenVersion int
	ProfileID    string
}

func ValidateJWT(signedToken, tokenSecret string) (ValidateJWTResults, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	}

	token, err := jwt.ParseWithClaims(signedToken, &customClaims{}, keyFunc)
	if err != nil {
		return ValidateJWTResults{}, fmt.Errorf("token parsing failed: %w", err)
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok {
		return ValidateJWTResults{}, errors.New("unknown claims type")
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return ValidateJWTResults{}, fmt.Errorf("error getting the token's subject: %w", err)
	}

	return ValidateJWTResults{
		TokenVersion: claims.TokenVersion,
		ProfileID:    subject,
	}, nil
}
