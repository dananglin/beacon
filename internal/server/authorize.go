// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/discovery"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
)

type MismatchedProfileIDError struct {
	authenticatedProfileID string
	profileIDInRequest     string
}

func (e MismatchedProfileIDError) Error() string {
	return "the authenticated profile ID (" +
		e.authenticatedProfileID +
		") does not match the profile ID in the authorize request (" +
		e.profileIDInRequest +
		")"
}

type authorizeRequest struct {
	ClientID            string
	CodeChallenge       string
	CodeChallengeMethod string
	Me                  string
	RedirectURI         string
	ResponseType        string
	Scope               []string
	State               string
}

func (s *Server) authorize(writer http.ResponseWriter, request *http.Request, profileID string) {
	cookie, err := request.Cookie(s.indieauthCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
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

	var authReq authorizeRequest
	if err := utilities.GobDecode(cookie.Value, &authReq); err != nil {
		sendServerError(
			writer,
			fmt.Errorf("error decoding the client's authorize request from cookie: %w", err),
		)

		return
	}

	// Ensure that the profile ID in the client's authorisation request matches the authenticated
	// profile ID.
	if authReq.Me != profileID {
		sendClientError(
			writer,
			http.StatusUnauthorized,
			MismatchedProfileIDError{
				authenticatedProfileID: profileID,
				profileIDInRequest:     authReq.Me,
			},
		)

		return
	}

	// Ensure that the client ID is a valid URL
	if err := utilities.ValidateClientURL(authReq.ClientID); err != nil {
		sendClientError(
			writer,
			http.StatusUnauthorized,
			fmt.Errorf(
				"the client ID (%s) does not appear to be a valid URL: %w",
				authReq.ClientID,
				err,
			),
		)

		return
	}

	// Fetch the client's metadata and use it to validate the authorization request
	clientMetadata, err := discovery.FetchClientMetadata(request.Context(), authReq.ClientID)
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf(
				"error fetching the client's metadata: %w",
				err,
			),
		)

		return
	}

	if err := discovery.ValidateClientMetadata(
		clientMetadata,
		authReq.ClientID,
		authReq.RedirectURI,
	); err != nil {
		sendClientError(
			writer,
			http.StatusUnauthorized,
			fmt.Errorf(
				"error validating the client's authorization request: %w",
				err,
			),
		)

		return
	}

	fmt.Fprintln(writer, "Good checkpoint reached! :)")
}

func (s *Server) authorizeRedirectToLogin(writer http.ResponseWriter, request *http.Request) {
	profileID := request.URL.Query().Get("me")
	scope := strings.Split(request.URL.Query().Get("scope"), " ")

	authRequest := authorizeRequest{
		ClientID:            request.URL.Query().Get("client_id"),
		CodeChallenge:       request.URL.Query().Get("code_challenge"),
		CodeChallengeMethod: request.URL.Query().Get("code_challenge_method"),
		Me:                  profileID,
		RedirectURI:         request.URL.Query().Get("redirect_uri"),
		ResponseType:        request.URL.Query().Get("response_type"),
		Scope:               scope,
		State:               request.URL.Query().Get("state"),
	}

	encodedString, err := utilities.GobEncode(authRequest)
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf("error encoding the client authorized request to gob data: %w", err),
		)
	}

	expiry := 10 * time.Minute

	cookie := http.Cookie{
		Name:     s.indieauthCookieName,
		Value:    encodedString,
		Path:     s.indieauthEndpoint,
		MaxAge:   int(expiry.Seconds()),
		Quoted:   false,
		Domain:   s.domainName,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(writer, &cookie)

	redirectURL := fmt.Sprintf(
		"/profile/login?login_type=%s&profile_id=%s",
		loginTypeIndieauth,
		url.QueryEscape(profileID),
	)

	http.Redirect(writer, request, redirectURL, http.StatusSeeOther)
}
