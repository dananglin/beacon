// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/auth"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
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

type MissingQueryValueError struct {
	parameter string
}

func (e MissingQueryValueError) Error() string {
	return "the value for '" + e.parameter + "' is missing from the URL's query"
}

func (s *Server) authorize(writer http.ResponseWriter, request *http.Request, profileID string) {
	authReq, err := newClientAuthRequest(s.indieauthCookieName, request)
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf("error getting the client's authorization request: %w", err),
		)

		return
	}

	// Ensure that the profile ID in the client's authorisation request matches the authenticated
	// profile ID.
	if authReq.Me != "" && authReq.Me != profileID {
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

	clientRequest := struct {
		ClientID          string
		ClientName        string
		ClientURI         string
		ClientRedirectURI string
		ProfileID         string
		AcceptURI         string
		RejectURI         string
		Scopes            []string
	}{
		ClientID:          clientMetadata.ClientID,
		ClientName:        clientMetadata.ClientName,
		ClientURI:         clientMetadata.ClientURI,
		ProfileID:         profileID,
		ClientRedirectURI: authReq.RedirectURI,
		AcceptURI:         s.authPath + "/accept",
		RejectURI:         s.authPath + "/reject",
		Scopes:            authReq.Scope,
	}

	generateAndSendHTMLResponse(
		writer,
		"authorization",
		http.StatusOK,
		clientRequest,
	)
}

func (s *Server) authorizeRedirectToLogin(writer http.ResponseWriter, request *http.Request) {
	profileID := request.URL.Query().Get("me")

	authRequest, err := newClientAuthRequestFromQuery(request.URL.Query())
	if err != nil {
		sendClientError(
			writer,
			http.StatusBadRequest,
			fmt.Errorf("error creating the client authorization request from query: %w", err),
		)

		return
	}

	encodedAuthRequest, err := utilities.GobEncode(authRequest)
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf("error encoding the client authorized request to gob data: %w", err),
		)

		return
	}

	expiry := 10 * time.Minute

	cookie := http.Cookie{
		Name:     s.indieauthCookieName,
		Value:    encodedAuthRequest,
		Path:     "/",
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

type clientRequestData struct {
	ClientID            string
	CodeChallenge       string
	CodeChallengeMethod string
	RedirectURI         string
	Scopes              []string
	Me                  string
}

func (s *Server) authorizeAccept(writer http.ResponseWriter, request *http.Request, profileID string) {
	authReq, err := newClientAuthRequest(s.indieauthCookieName, request)
	if err != nil {
		sendClientError(
			writer,
			http.StatusUnauthorized,
			fmt.Errorf("error getting the client's authorization request: %w", err),
		)

		return
	}

	// Create the authorization code
	authCodeBytes := make([]byte, 32)

	if _, err := rand.Read(authCodeBytes); err != nil {
		sendServerError(
			writer,
			fmt.Errorf("unable to create random bytes: %w", err),
		)

		return
	}

	authCode := hex.EncodeToString(authCodeBytes)

	// Data associated with the authorization code
	authResp := clientRequestData{
		ClientID:            authReq.ClientID,
		CodeChallenge:       authReq.CodeChallenge,
		CodeChallengeMethod: authReq.CodeChallengeMethod,
		RedirectURI:         authReq.RedirectURI,
		Scopes:              authReq.Scope,
		Me:                  profileID,
	}

	buffer := new(bytes.Buffer)
	if err := gob.NewEncoder(buffer).Encode(authResp); err != nil {
		sendServerError(
			writer,
			fmt.Errorf("unable to encode the data: %w", err),
		)

		return
	}

	// Save code and associated data to the server's cache
	s.cache.Add(
		authCode,
		buffer.Bytes(),
		time.Now().Add(1*time.Minute),
	)

	// Construct the authorization response and set it to the header below.
	redirectURL := fmt.Sprintf(
		"%s?code=%s&state=%s&iss=%s",
		authReq.RedirectURI,
		url.QueryEscape(authCode),
		url.QueryEscape(authReq.State),
		url.QueryEscape(s.issuer),
	)

	writer.Header().Add("HX-Redirect", redirectURL)
}

func (s *Server) authorizeReject(writer http.ResponseWriter, request *http.Request, _ string) {
	authReq, err := newClientAuthRequest(s.indieauthCookieName, request)
	if err != nil {
		sendClientError(
			writer,
			http.StatusUnauthorized,
			fmt.Errorf("error getting the client's authorization request: %w", err),
		)

		return
	}

	redirectURL := fmt.Sprintf(
		"%s?error=access_denied&state=%s",
		authReq.RedirectURI,
		url.QueryEscape(authReq.State),
	)

	writer.Header().Add("HX-Redirect", redirectURL)
}

func (s *Server) profileExchange(writer http.ResponseWriter, request *http.Request, data clientRequestData) {
	// Get the profile information if requested.
	profile := make(map[string]string)

	if slices.Contains(data.Scopes, "profile") {
		// Get the profile information from the database
		info, err := database.GetProfileInformation(s.boltdb, data.Me)
		if err != nil {
			sendServerError(
				writer,
				fmt.Errorf("unable to get the profile information for %q: %w", data.Me, err),
			)

			return
		}

		profile = map[string]string{
			"name":  info.Name,
			"url":   info.URL,
			"photo": info.PhotoURL,
		}

		if slices.Contains(data.Scopes, "email") {
			profile["email"] = info.Email
		}
	}

	// Construct the JSON response and send it back to the client
	response := struct {
		Me      string            `json:"me"`
		Profile map[string]string `json:"profile,omitempty"`
	}{
		Me:      data.Me,
		Profile: profile,
	}

	sendJSONResponse(writer, http.StatusOK, response)
}

func (s *Server) tokenExchange(writer http.ResponseWriter, request *http.Request, data clientRequestData) {
	var err error

	// Create the access token.
	// If there are no requested scopes then the access token won't be created.
	bearerToken := ""
	if len(data.Scopes) > 0 {
		bearerToken, err = auth.CreateBearerToken()
		if err != nil {
			sendServerError(
				writer,
				fmt.Errorf("unable to create the bearer token: %w", err),
			)

			return
		}
	}

	// Get the profile information if requested.
	profile := make(map[string]string)

	if slices.Contains(data.Scopes, "profile") {
		// Get the profile information from the database
		info, err := database.GetProfileInformation(s.boltdb, data.Me)
		if err != nil {
			sendServerError(
				writer,
				fmt.Errorf("unable to get the profile information for %q: %w", data.Me, err),
			)

			return
		}

		profile = map[string]string{
			"name":  info.Name,
			"url":   info.URL,
			"photo": info.PhotoURL,
		}

		if slices.Contains(data.Scopes, "email") {
			profile["email"] = info.Email
		}
	}

	// Construct the JSON response and send it back to the client.
	response := struct {
		AccessToken string            `json:"access_token"`
		TokenType   string            `json:"token_type"`
		Scope       string            `json:"scope"`
		Me          string            `json:"me"`
		Profile     map[string]string `json:"profile,omitempty"`
	}{
		AccessToken: bearerToken,
		TokenType:   "Bearer",
		Scope:       strings.Join(data.Scopes, " "),
		Me:          data.Me,
		Profile:     profile,
	}

	sendJSONResponse(writer, http.StatusOK, response)
}

type clientAuthRequest struct {
	ClientID            string
	CodeChallenge       string
	CodeChallengeMethod string
	Me                  string
	RedirectURI         string
	ResponseType        string
	Scope               []string
	State               string
}

func newClientAuthRequest(cookieName string, request *http.Request) (clientAuthRequest, error) {
	cookie, err := request.Cookie(cookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			authReq, err := newClientAuthRequestFromQuery(request.URL.Query())
			if err != nil {
				return clientAuthRequest{}, fmt.Errorf(
					"unable to create the client authorization request: %w",
					err,
				)
			}

			return authReq, nil
		}

		return clientAuthRequest{}, fmt.Errorf("unable to retrieve the cookie: %w", err)
	}

	var authReq clientAuthRequest
	if err := utilities.GobDecode(cookie.Value, &authReq); err != nil {
		return clientAuthRequest{}, fmt.Errorf(
			"error decoding the client's authorize request from cookie: %w",
			err,
		)
	}

	return authReq, nil
}

func newClientAuthRequestFromQuery(queryValues url.Values) (clientAuthRequest, error) {
	const (
		clientID            = "client_id"
		codeChallenge       = "code_challenge"
		codeChallengeMethod = "code_challenge_method"
		me                  = "me"
		redirectURI         = "redirect_uri"
		responseType        = "response_type"
		scope               = "scope"
		state               = "state"
	)

	required := []string{
		clientID,
		codeChallenge,
		codeChallengeMethod,
		redirectURI,
		responseType,
		scope,
		state,
	}

	for ind := range required {
		if !queryValues.Has(required[ind]) {
			return clientAuthRequest{}, MissingQueryValueError{parameter: required[ind]}
		}
	}

	request := clientAuthRequest{
		ClientID:            queryValues.Get(clientID),
		CodeChallenge:       queryValues.Get(codeChallenge),
		CodeChallengeMethod: queryValues.Get(codeChallengeMethod),
		Me:                  queryValues.Get(me),
		RedirectURI:         queryValues.Get(redirectURI),
		ResponseType:        queryValues.Get(responseType),
		Scope:               strings.Split(queryValues.Get(scope), " "),
		State:               queryValues.Get(state),
	}

	return request, nil
}
