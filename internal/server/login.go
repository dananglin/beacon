// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/auth"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
)

const (
	loginTypeProfile = "profile"
)

type formLogin struct {
	AuthenticationFailure bool
	ProfileID             string
	Password              string
	FieldErrors           map[string]string
	LoginType             string
}

func (f *formLogin) validate() bool {
	f.FieldErrors = make(map[string]string)

	if strings.TrimSpace(f.ProfileID) == "" {
		f.FieldErrors["ProfileID"] = "This field cannot be blank"
	}

	if strings.TrimSpace(f.Password) == "" {
		f.FieldErrors["Password"] = "This field cannot be blank"
	}

	return len(f.FieldErrors) == 0
}

func (s *Server) getLoginForm(writer http.ResponseWriter, request *http.Request) {
	loginType := request.URL.Query().Get("login_type")
	if loginType == "" {
		loginType = loginTypeProfile
	}

	form := formLogin{
		AuthenticationFailure: false,
		ProfileID:             "",
		Password:              "",
		FieldErrors:           make(map[string]string),
		LoginType:             loginType,
	}

	generateAndSendHTMLResponse(
		writer,
		"login",
		http.StatusOK,
		form,
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

	form := formLogin{
		AuthenticationFailure: false,
		ProfileID:             request.PostFormValue("profileID"),
		Password:              request.PostFormValue("password"),
		FieldErrors:           make(map[string]string),
		LoginType:             request.PostFormValue("loginType"),
	}

	if validForm := form.validate(); !validForm {
		generateAndSendHTMLResponse(
			writer,
			"login",
			http.StatusUnprocessableEntity,
			form,
		)

		return
	}

	profileID, err := utilities.ValidateProfileURL(form.ProfileID)
	if err != nil {
		form.AuthenticationFailure = true

		generateAndSendHTMLResponse(
			writer,
			"login",
			http.StatusUnauthorized,
			form,
		)

		return
	}

	password := form.Password

	profile, err := database.GetProfile(s.boltdb, profileID)
	if err != nil {
		form.AuthenticationFailure = true

		generateAndSendHTMLResponse(
			writer,
			"login",
			http.StatusUnauthorized,
			form,
		)

		return
	}

	if err := auth.CheckPasswordHash(profile.HashedPassword, password); err != nil {
		form.AuthenticationFailure = true

		generateAndSendHTMLResponse(
			writer,
			"login",
			http.StatusUnauthorized,
			form,
		)

		return
	}

	expiry := 1 * time.Hour

	token, err := auth.CreateJWT(profileID, s.jwtSecret, profile.TokenVersion, expiry)
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf("error creating the JWT: %w", err),
		)

		return
	}

	cookie := http.Cookie{
		Name:     s.cookieName,
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
		loginTypeProfile: "/profile/overview",
	}

	redirect, ok := redirectMap[form.LoginType]
	if !ok {
		sendClientError(
			writer,
			http.StatusBadRequest,
			fmt.Errorf("unrecognised login type: %s", form.LoginType),
		)
	}

	http.Redirect(writer, request, redirect, http.StatusSeeOther)
}
