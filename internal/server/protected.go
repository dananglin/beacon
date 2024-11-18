// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/auth"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
)

func (s *Server) protected(
	next func(writer http.ResponseWriter, request *http.Request, profileID string),
	redirectToLogin func(writer http.ResponseWriter, request *http.Request),
) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cookie, err := request.Cookie(s.jwtCookieName)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				if redirectToLogin != nil {
					redirectToLogin(writer, request)

					return
				}

				sendClientError(
					writer,
					http.StatusUnauthorized,
					err,
				)

				return
			}

			sendServerError(
				writer,
				fmt.Errorf("error getting cookie: %w", err),
			)

			return
		}

		token := cookie.Value

		data, err := auth.ValidateJWT(token, s.jwtSecret)
		if err != nil {
			sendClientError(
				writer,
				http.StatusUnauthorized,
				fmt.Errorf("token validation error: %w", err),
			)

			return
		}

		profileTokenVersion, err := database.GetProfileTokenVersion(s.boltdb, data.ProfileID)
		if err != nil {
			profileNotExistErr := database.ProfileNotExistError{}
			if errors.As(err, &profileNotExistErr) {
				sendClientError(
					writer,
					http.StatusUnauthorized,
					err,
				)

				return
			}

			sendServerError(
				writer,
				fmt.Errorf("error getting the profile's token version: %w", err),
			)

			return
		}

		if profileTokenVersion != data.TokenVersion {
			sendClientError(
				writer,
				http.StatusUnauthorized,
				errors.New("invalid token"),
			)

			return
		}

		// Set the "Cache-Control: no-store" header so that protected pages
		// are not stored in the browser's cache.
		writer.Header().Add("Cache-Control", "no-store")

		next(writer, request, data.ProfileID)
	})
}

func (s *Server) exchangeAuthorization(
	exchange func(writer http.ResponseWriter, request *http.Request, data clientRequestData),
) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseForm(); err != nil {
			sendClientError(
				writer,
				http.StatusBadRequest,
				fmt.Errorf("error parsing the form: %w", err),
			)

			return
		}

		var (
			grantType    = request.PostFormValue("grant_type")
			code         = request.PostFormValue("code")
			clientID     = request.PostFormValue("client_id")
			redirectURI  = request.PostFormValue("redirect_uri")
			codeVerifier = request.PostFormValue("code_verifier")
		)

		// The grant type must be "authorization_code"
		if grantType == "" {
			sendClientError(
				writer,
				http.StatusUnprocessableEntity,
				errors.New("the required parameter 'grant_type' is missing"),
			)

			return
		}

		if grantType != "authorization_code" {
			sendClientError(
				writer,
				http.StatusUnprocessableEntity,
				fmt.Errorf("unsupported grant_type: %q", grantType),
			)

			return
		}

		// Using the code to get the associated data from the server's cache
		cacheEntry, exists := s.cache.Get(code)
		if !exists {
			sendClientError(
				writer,
				http.StatusUnauthorized,
				errors.New("invalid authorization code: the code is not present in the cache"),
			)

			return
		}

		if cacheEntry.Expired() {
			sendClientError(
				writer,
				http.StatusUnauthorized,
				errors.New("invalid authorization code: the code has expired"),
			)

			return
		}

		var data clientRequestData

		buffer := bytes.NewBuffer(cacheEntry.Value())
		if err := gob.NewDecoder(buffer).Decode(&data); err != nil {
			sendServerError(
				writer,
				fmt.Errorf("unable to decode the data from the cache: %w", err),
			)

			return
		}

		// The client ID must match
		if clientID != data.ClientID {
			sendClientError(
				writer,
				http.StatusUnauthorized,
				errors.New("mismatched client ID"),
			)

			return
		}

		// The redirect URI must match
		if redirectURI != data.RedirectURI {
			sendClientError(
				writer,
				http.StatusUnauthorized,
				errors.New("mismatched redirect URI"),
			)

			return
		}

		// Verify the code verifier
		if err := auth.VerifyAuthorizationCode(
			data.CodeChallengeMethod,
			data.CodeChallenge,
			codeVerifier,
		); err != nil {
			sendClientError(
				writer,
				http.StatusUnauthorized,
				fmt.Errorf("error verifying the code verifier: %w", err),
			)

			return
		}

		// The client is now authorized to complete the required exchange.
		exchange(writer, request, data)

		// Delete the code and associated data from the cache
		// to ensure that is doesn't get used again.
		s.cache.Delete(code)
	})
}
