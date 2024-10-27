// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"errors"
	"fmt"
	"net/http"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/auth"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
)

func (s *Server) protected(handler func(writer http.ResponseWriter, request *http.Request, profileID string)) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cookie, err := request.Cookie(s.cookieName)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				sendClientError(
					writer,
					http.StatusBadRequest,
					err,
				)
			} else {
				sendServerError(
					writer,
					fmt.Errorf("error getting cookie: %w", err),
				)
			}

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

		handler(writer, request, data.ProfileID)
	})
}
