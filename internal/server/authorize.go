// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
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

func (s *Server) authorize(writer http.ResponseWriter, request *http.Request, profileID string) {
	encodedState := request.URL.Query().Get("state")

	authReq, err := s.getClientAuthRequest(&encodedState, request.URL.Query())
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
	clientMetadata, err := discovery.FetchClientMetadata(request.Context(), authReq.ClientID, s.issuer)
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

	dataForHTML := struct {
		ClientID          string
		ClientName        string
		ClientURI         string
		ClientRedirectURI string
		ProfileID         string
		AcceptURI         string
		RejectURI         string
		State             string
		Scopes            []string
	}{
		ClientID:          clientMetadata.ClientID,
		ClientName:        clientMetadata.ClientName,
		ClientURI:         clientMetadata.ClientURI,
		ProfileID:         profileID,
		ClientRedirectURI: authReq.RedirectURI,
		AcceptURI:         s.authPath + "/accept",
		RejectURI:         s.authPath + "/reject",
		State:             encodedState,
		Scopes:            authReq.Scope,
	}

	generateAndSendHTMLResponse(
		writer,
		"authorization",
		http.StatusOK,
		dataForHTML,
	)
}

func (s *Server) authorizeRedirectToLogin(writer http.ResponseWriter, request *http.Request) {
	authRequest, err := newClientAuthRequestFromQuery(request.URL.Query())
	if err != nil {
		sendClientError(
			writer,
			http.StatusBadRequest,
			fmt.Errorf("error creating the client authorization request from query: %w", err),
		)

		return
	}

	encodedState, err := s.addClientAuthRequestToCache(authRequest)
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf("error adding the client auth request to cache: %w", err),
		)
	}

	profileID := authRequest.Me

	redirectURL := fmt.Sprintf(
		"/profile/login?login_type=%s&profile_id=%s&state=%s",
		loginTypeIndieauth,
		url.QueryEscape(profileID),
		encodedState,
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
	if err := request.ParseForm(); err != nil {
		sendClientError(
			writer,
			http.StatusBadRequest,
			fmt.Errorf("error parsing the form: %w", err),
		)

		return
	}

	encodedState := request.PostFormValue("state")

	authReq, err := s.getClientAuthRequest(&encodedState, request.URL.Query())
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

	authRespBytes, err := utilities.GobEncode(authResp)
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf("unable to encode the data: %w", err),
		)

		return
	}

	// Save code and associated data to the server's cache
	s.cache.Add(
		authCode,
		authRespBytes,
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
	if err := request.ParseForm(); err != nil {
		sendClientError(
			writer,
			http.StatusBadRequest,
			fmt.Errorf("error parsing the form: %w", err),
		)

		return
	}

	encodedState := request.PostFormValue("state")

	authReq, err := s.getClientAuthRequest(&encodedState, request.URL.Query())
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

func (s *Server) profileExchange(writer http.ResponseWriter, data clientRequestData) {
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

func (s *Server) tokenExchange(writer http.ResponseWriter, data clientRequestData) {
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

// getClientAuthRequest attempts to retrieve the client's authorization request from cache. If this is not found
// the it attempts to create the request from the HTTP request's query.
func (s *Server) getClientAuthRequest(encodedState *string, queryValues url.Values) (clientAuthRequest, error) {
	authReq, err := s.getClientAuthRequestFromCache(*encodedState)
	if err == nil {
		return authReq, nil
	}

	notFoundErr := StateKeyNotFoundInCacheError{}
	if !errors.As(err, &notFoundErr) {
		return clientAuthRequest{}, fmt.Errorf(
			"unexpected error received after attempting to get the client authorization request from cache: %w",
			err,
		)
	}

	authReq, err = newClientAuthRequestFromQuery(queryValues)
	if err != nil {
		return clientAuthRequest{}, fmt.Errorf(
			"error creating the client authorization request: %w",
			err,
		)
	}

	encodedStateFromCache, err := s.addClientAuthRequestToCache(authReq)
	if err != nil {
		return clientAuthRequest{}, fmt.Errorf(
			"error adding the client authorization request to the cache: %w",
			err,
		)
	}

	*encodedState = encodedStateFromCache

	return authReq, nil
}

// newClientAuthRequestFromQuery extracts the client's authorization request from the HTTP request's
// query. An error is returned if one of the required parameters is missing from the query.
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
		state,
	}

	for ind := range required {
		if !queryValues.Has(required[ind]) {
			return clientAuthRequest{}, MissingQueryValueError{parameter: required[ind]}
		}
	}

	scopeStr := queryValues.Get(scope)

	scopes := make([]string, 0)

	if scopeStr != "" {
		scopes = strings.Split(scopeStr, " ")
	}

	request := clientAuthRequest{
		ClientID:            queryValues.Get(clientID),
		CodeChallenge:       queryValues.Get(codeChallenge),
		CodeChallengeMethod: queryValues.Get(codeChallengeMethod),
		Me:                  queryValues.Get(me),
		RedirectURI:         queryValues.Get(redirectURI),
		ResponseType:        queryValues.Get(responseType),
		Scope:               scopes,
		State:               queryValues.Get(state),
	}

	return request, nil
}

// addClientAuthRequestToCache adds the client's authorize request to cache. The value of the state
// is encoded using Base64 URL encoding and is used as the key to access the associated request.
// After the data is added to the cache, the encoded state value is returned to the caller.
func (s *Server) addClientAuthRequestToCache(request clientAuthRequest) (string, error) {
	encodedState := base64.URLEncoding.EncodeToString([]byte(request.State))

	if _, exists := s.cache.Get(encodedState); exists {
		return "", ExistingStateKeyInCacheError{encodedState: encodedState}
	}

	requestBytes, err := utilities.GobEncode(request)
	if err != nil {
		return "", fmt.Errorf("error gob encoding the client auth request: %w", err)
	}

	s.cache.Add(encodedState, requestBytes, time.Now().Add(10*time.Minute))

	return encodedState, nil
}

// getClientAuthRequestFromCache attempts to retrieve the client's authorization request from the cache.
func (s *Server) getClientAuthRequestFromCache(encodedState string) (clientAuthRequest, error) {
	cachedRequest, exists := s.cache.Get(encodedState)
	if !exists {
		return clientAuthRequest{}, StateKeyNotFoundInCacheError{encodedState: encodedState}
	}

	var request clientAuthRequest

	if err := utilities.GobDecode(bytes.NewBuffer(cachedRequest.Value()), &request); err != nil {
		return clientAuthRequest{}, fmt.Errorf("error gob decoding the client auth request: %w", err)
	}

	return request, nil
}
