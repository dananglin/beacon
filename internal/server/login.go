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

func (s *Server) getLoginForm(writer http.ResponseWriter, _ *http.Request) {
	form := formLogin{
		AuthenticationFailure: false,
		ProfileID:             "",
		Password:              "",
		FieldErrors:           make(map[string]string),
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
		Name:     "beacon_is_great",
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

	http.Redirect(writer, request, "/profile/login/confirmation", http.StatusSeeOther)
}

type formLogin struct {
	AuthenticationFailure bool
	ProfileID             string
	Password              string
	FieldErrors           map[string]string
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
