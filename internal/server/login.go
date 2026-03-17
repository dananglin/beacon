// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
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

	"codeflow.dananglin.me.uk/apollo/beacon/internal/auth"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/info"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
)

type loginPage struct {
	ProfileID      string
	LoginType      string
	State          string
	Title          string
	FailureMessage string
	FieldErrors    map[string]string
}

type loginForm struct {
	profileID string
	password  string
	loginType string
	state     string
}

func (f *loginForm) valid() (map[string]string, bool) {
	fieldErrors := make(map[string]string)

	if strings.TrimSpace(f.profileID) == "" {
		fieldErrors["ProfileID"] = "This field cannot be blank"
	}

	if strings.TrimSpace(f.password) == "" {
		fieldErrors["Password"] = "This field cannot be blank"
	}

	return fieldErrors, len(fieldErrors) == 0
}

func (s *Server) getLoginPage(writer http.ResponseWriter, request *http.Request) {
	query, err := url.ParseQuery(request.URL.RawQuery)
	if err != nil {
		sendClientError(
			writer,
			http.StatusBadRequest,
			fmt.Errorf("error parsing the query string: %w", err),
		)
	}

	loginType := query.Get(qKeyLoginType)
	if loginType != loginTypeIndieauth {
		s.sendHTMLResponse(
			writer,
			"login",
			http.StatusOK,
			loginPage{
				ProfileID:      "",
				LoginType:      loginTypeProfile,
				State:          "",
				Title:          loginPageTitle(),
				FailureMessage: "",
				FieldErrors:    map[string]string{},
			},
			nil,
			nil,
		)

		return
	}

	profileID := query.Get(qKeyProfileID)
	state := query.Get(qKeyState)

	s.sendHTMLResponse(
		writer,
		"login",
		http.StatusOK,
		loginPage{
			ProfileID:      profileID,
			LoginType:      loginType,
			State:          state,
			Title:          loginPageTitle(),
			FailureMessage: "",
			FieldErrors:    map[string]string{},
		},
		nil,
		nil,
	)
}

func (s *Server) authenticate(writer http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		sendClientError(
			writer,
			http.StatusBadRequest,
			fmt.Errorf("error parsing the form: %w", err),
		)

		return
	}

	form := loginForm{
		profileID: request.PostFormValue("profileID"),
		password:  request.PostFormValue("password"),
		loginType: request.PostFormValue("loginType"),
		state:     request.PostFormValue("state"),
	}

	fieldErrors, valid := form.valid()
	if !valid {
		page := loginPage{
			ProfileID:      form.profileID,
			LoginType:      form.loginType,
			State:          form.state,
			Title:          loginPageTitle(),
			FailureMessage: "Invalid form",
			FieldErrors:    fieldErrors,
		}

		s.sendHTMLResponse(
			writer,
			"login",
			http.StatusUnprocessableEntity,
			page,
			errors.New("invalid form"),
			nil,
		)

		return
	}

	profileID, err := utilities.ValidateAndCanonicalizeURL(form.profileID, false)
	if err != nil {
		s.sendHTMLResponse(
			writer,
			"login",
			http.StatusUnauthorized,
			loginPage{
				ProfileID:      form.profileID,
				LoginType:      form.loginType,
				State:          form.state,
				Title:          loginPageTitle(),
				FailureMessage: "The Profile ID or password is incorrect",
				FieldErrors:    fieldErrors,
			},
			fmt.Errorf("error validating and canonicalizing the profile ID: %w", err),
			nil,
		)

		return
	}

	exists, err := database.ProfileExists(s.boltdb, form.profileID)
	if err != nil {
		s.sendHTMLResponse(
			writer,
			"login",
			http.StatusInternalServerError,
			loginPage{
				ProfileID:      form.profileID,
				LoginType:      form.loginType,
				State:          form.state,
				Title:          loginPageTitle(),
				FailureMessage: "Unable to login",
				FieldErrors:    make(map[string]string),
			},
			nil,
			fmt.Errorf("error looking up the profile in the database: %w", err),
		)

		return
	}

	if !exists {
		s.sendHTMLResponse(
			writer,
			"login",
			http.StatusUnauthorized,
			loginPage{
				ProfileID:      form.profileID,
				LoginType:      form.loginType,
				State:          form.state,
				Title:          loginPageTitle(),
				FailureMessage: "The Profile ID or password is incorrect",
				FieldErrors:    make(map[string]string),
			},
			fmt.Errorf("unknown profile ID: %s", form.profileID),
			nil,
		)

		return
	}

	profile, err := database.GetProfile(s.boltdb, profileID)
	if err != nil {
		s.sendHTMLResponse(
			writer,
			"login",
			http.StatusInternalServerError,
			loginPage{
				ProfileID:      form.profileID,
				LoginType:      form.loginType,
				State:          form.state,
				Title:          loginPageTitle(),
				FailureMessage: "Unable to login",
				FieldErrors:    make(map[string]string),
			},
			nil,
			fmt.Errorf("error retrieving the profile from the database: %w", err),
		)

		return
	}

	err = auth.CheckPasswordHash(profile.HashedPassword, form.password)
	if err != nil {
		s.sendHTMLResponse(
			writer,
			"login",
			http.StatusUnauthorized,
			loginPage{
				ProfileID:      form.profileID,
				LoginType:      form.loginType,
				State:          form.state,
				Title:          loginPageTitle(),
				FailureMessage: "The Profile ID or password is incorrect",
				FieldErrors:    make(map[string]string),
			},
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
			"login",
			http.StatusInternalServerError,
			loginPage{
				ProfileID:      form.profileID,
				LoginType:      form.loginType,
				State:          form.state,
				Title:          loginPageTitle(),
				FailureMessage: "Unable to login",
				FieldErrors:    make(map[string]string),
			},
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
		loginTypeIndieauth: fmt.Sprintf("%s?state=%s", s.authPath, form.state),
	}

	redirectURL, ok := redirectMap[form.loginType]
	if !ok {
		s.sendHTMLResponse(
			writer,
			"login",
			http.StatusBadRequest,
			loginPage{
				ProfileID:      form.profileID,
				LoginType:      form.loginType,
				State:          form.state,
				Title:          loginPageTitle(),
				FailureMessage: "Unable to login",
				FieldErrors:    make(map[string]string),
			},
			fmt.Errorf("unrecognised login type: %s", form.loginType),
			nil,
		)

		return
	}

	http.Redirect(writer, request, redirectURL, http.StatusSeeOther)
}

func (s *Server) logout(writer http.ResponseWriter, request *http.Request, profileID string) {
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
