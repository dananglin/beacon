// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/auth"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/info"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
)

type loginPage struct {
	ProfileID string
	LoginType string
	State     string
	Title     string
}

type loginForm struct {
	profileID string
	password  string
	loginType string
	state     string
}

func (f *loginForm) validate() error {
	if strings.TrimSpace(f.profileID) == "" {
		return formValidationError{reason: "the profile ID is not set"}
	}

	if strings.TrimSpace(f.password) == "" {
		return formValidationError{reason: "the password is not set"}
	}

	return nil
}

func (s *Server) getLoginPage(writer http.ResponseWriter, request *http.Request) {
	query, err := url.ParseQuery(request.URL.RawQuery)
	if err != nil {
		sendClientError(
			writer,
			http.StatusBadRequest,
			fmt.Errorf("error parsing the query string: %w", err),
		)

		return
	}

	loginType := query.Get(qKeyLoginType)
	if loginType != loginTypeIndieauth {
		s.sendHTMLResponseWithTemplate(
			writer,
			"login",
			http.StatusOK,
			loginPage{
				ProfileID: "",
				LoginType: loginTypeProfile,
				State:     "",
				Title:     loginPageTitle(),
			},
			nil,
			nil,
		)

		return
	}

	profileID := query.Get(qKeyProfileID)
	state := query.Get(qKeyState)

	s.sendHTMLResponseWithTemplate(
		writer,
		"login",
		http.StatusOK,
		loginPage{
			ProfileID: profileID,
			LoginType: loginType,
			State:     state,
			Title:     loginPageTitle(),
		},
		nil,
		nil,
	)
}

func (s *Server) authenticate(writer http.ResponseWriter, request *http.Request) {
	form := loginForm{
		profileID: request.PostFormValue("profileID"),
		password:  request.PostFormValue("password"),
		loginType: request.PostFormValue("loginType"),
		state:     request.PostFormValue("state"),
	}

	err := form.validate()
	if err != nil {
		s.sendHTMLResponse(
			writer,
			fmt.Appendf([]byte{}, responseFailureFmt, "The Profile ID or password is incorrect"),
			http.StatusUnauthorized,
			err,
			nil,
		)

		return
	}

	profileID, err := utilities.ValidateAndCanonicalizeURL(form.profileID, false)
	if err != nil {
		s.sendHTMLResponse(
			writer,
			fmt.Appendf([]byte{}, responseFailureFmt, "The Profile ID or password is incorrect"),
			http.StatusUnauthorized,
			fmt.Errorf("error validating and canonicalizing the profile ID: %w", err),
			nil,
		)

		return
	}

	exists, err := database.ProfileExists(s.boltdb, form.profileID)
	if err != nil {
		s.sendHTMLResponse(
			writer,
			fmt.Appendf([]byte{}, responseFailureFmt, "Unable to login"),
			http.StatusInternalServerError,
			nil,
			fmt.Errorf("error looking up the profile in the database: %w", err),
		)

		return
	}

	if !exists {
		s.sendHTMLResponse(
			writer,
			fmt.Appendf([]byte{}, responseFailureFmt, "The Profile ID or password is incorrect"),
			http.StatusUnauthorized,
			fmt.Errorf("unknown profile ID: %s", form.profileID),
			nil,
		)

		return
	}

	profile, err := database.GetProfile(s.boltdb, profileID)
	if err != nil {
		s.sendHTMLResponse(
			writer,
			fmt.Appendf([]byte{}, responseFailureFmt, "Unable to login"),
			http.StatusInternalServerError,
			nil,
			fmt.Errorf("error retrieving the profile from the database: %w", err),
		)

		return
	}

	err = auth.CheckPasswordHash(profile.HashedPassword, form.password)
	if err != nil {
		s.sendHTMLResponse(
			writer,
			fmt.Appendf([]byte{}, responseFailureFmt, "The Profile ID or password is incorrect"),
			http.StatusUnauthorized,
			fmt.Errorf("error validating the password: %w", err),
			nil,
		)

		return
	}

	expiry := 1 * time.Hour

	token, err := auth.CreateJWT(profileID, s.jwtSecret, profile.TokenVersion, expiry)
	if err != nil {
		s.sendHTMLResponse(
			writer,
			fmt.Appendf([]byte{}, responseFailureFmt, "Unable to login"),
			http.StatusInternalServerError,
			nil,
			fmt.Errorf("error creating the JWT: %w", err),
		)

		return
	}

	cookie := http.Cookie{
		Name:     s.jwtCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(expiry.Seconds()),
		Quoted:   false,
		Domain:   s.domainName,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(writer, &cookie)

	redirectMap := map[string]string{
		loginTypeProfile:   "/profile/overview",
		loginTypeIndieauth: fmt.Sprintf("%s?state=%s", pathAuth, form.state),
	}

	redirectURL, ok := redirectMap[form.loginType]
	if !ok {
		s.sendHTMLResponse(
			writer,
			fmt.Appendf([]byte{}, responseFailureFmt, "Unable to login"),
			http.StatusBadRequest,
			fmt.Errorf("unrecognised login type: %s", form.loginType),
			nil,
		)

		return
	}

	writer.Header().Set("Hx-Redirect", redirectURL)
}

func (s *Server) logout(writer http.ResponseWriter, _ *http.Request, profileID string) {
	if err := database.IncrementTokenVersion(s.boltdb, profileID); err != nil {
		sendServerError(
			writer,
			fmt.Errorf("error incrementing the profile's token version: %w", err),
		)

		return
	}

	writer.Header().Set("Hx-Redirect", "/profile/login")
}

func loginPageTitle() string {
	return "Sign in - " + info.ApplicationTitledName
}
