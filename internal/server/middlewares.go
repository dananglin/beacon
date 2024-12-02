// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/auth"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
)

// entrypoint is the middleware that acts as the entry point of all requests. The entrypoint
// assigns each request with a unique ID for troubleshooting and writes an access log when each
// request is completed.
func (s *Server) entrypoint(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !s.dbInitialized {
			http.Redirect(writer, request, "/setup", http.StatusSeeOther)

			return
		}

		requestID := "UNKNOWN"
		id := make([]byte, 16)

		if _, err := rand.Read(id); err != nil {
			slog.Error("unable to create the request ID.", "error", err.Error())
		} else {
			requestID = hex.EncodeToString(id)
		}

		writer.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(writer, request)

		// TODO: Write access log
	})
}

// profileAuthorization is a middleware that performs profile authorization before calling
// the profile handler. If the cookie storing the authorization information is missing then
// the user is redirected to the login screen.
func (s *Server) profileAuthorization(next profileHandlerFunc, redirectToLogin http.HandlerFunc) http.Handler {
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
			if redirectToLogin != nil {
				redirectToLogin(writer, request)

				return
			}

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
			if redirectToLogin != nil {
				redirectToLogin(writer, request)

				return
			}

			sendClientError(
				writer,
				http.StatusUnauthorized,
				ErrInvalidProfileAccessToken,
			)

			return
		}

		// Set the "Cache-Control: no-store" header so that protected pages
		// are not stored in the browser's cache.
		writer.Header().Add("Cache-Control", "no-store")

		next(writer, request, data.ProfileID)
	})
}

func (s *Server) exchangeAuthorization(exchange exchangeHandlerFunc) http.Handler {
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

		// Regardless of the success or failure of the exchange authorization the code
		// and associated data is deleted from the cache to ensure that is doesn't get
		// used again.
		defer s.cache.Delete(code)

		// The grant type must be "authorization_code"
		if grantType == "" {
			sendClientError(
				writer,
				http.StatusUnprocessableEntity,
				ErrMissingGrantType,
			)

			return
		}

		if grantType != "authorization_code" {
			sendClientError(
				writer,
				http.StatusUnprocessableEntity,
				UnsupportedGrantTypeError{grantType: grantType},
			)

			return
		}

		// Using the code to get the associated data from the server's cache
		cacheEntry, exists := s.cache.Get(code)
		if !exists {
			sendClientError(
				writer,
				http.StatusUnauthorized,
				ErrMissingAuthorizationCode,
			)

			return
		}

		if cacheEntry.Expired() {
			sendClientError(
				writer,
				http.StatusUnauthorized,
				ErrExpiredAuthorizationCode,
			)

			return
		}

		var initialClientAuthReq clientRequestData

		if err := utilities.GobDecode(bytes.NewBuffer(cacheEntry.Value()), &initialClientAuthReq); err != nil {
			sendServerError(
				writer,
				fmt.Errorf("unable to decode the data from the cache: %w", err),
			)

			return
		}

		// The client ID must match
		canonicalizedClientID, err := utilities.ValidateAndCanonicalizeURL(clientID, true)
		if err != nil {
			sendClientError(
				writer,
				http.StatusUnauthorized,
				fmt.Errorf("error canonicalizing the client ID: %w", err),
			)

			return
		}

		if canonicalizedClientID != initialClientAuthReq.ClientID {
			sendClientError(
				writer,
				http.StatusUnauthorized,
				MismatchedClientIDError{
					exchangedClientID: clientID,
					initialClientID:   initialClientAuthReq.ClientID,
				},
			)

			return
		}

		// The redirect URI must match
		if redirectURI != initialClientAuthReq.RedirectURI {
			sendClientError(
				writer,
				http.StatusUnauthorized,
				MismatchedRedirectURIError{
					exchangedRedirectURI: redirectURI,
					initialRedirectURI:   initialClientAuthReq.RedirectURI,
				},
			)

			return
		}

		// Verify the code verifier
		if err := auth.VerifyAuthorizationCode(
			initialClientAuthReq.CodeChallengeMethod,
			initialClientAuthReq.CodeChallenge,
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
		exchange(writer, initialClientAuthReq)
	})
}

func neuter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasSuffix(request.URL.Path, "/") {
			sendClientError(
				writer,
				http.StatusNotFound,
				ErrInvalidFileserverPath,
			)

			return
		}

		next.ServeHTTP(writer, request)
	})
}
