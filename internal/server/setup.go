// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"unicode/utf8"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/auth"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
)

func (s *Server) setup(writer http.ResponseWriter, request *http.Request) {
	initialised, err := database.Initialised(s.boltdb)
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf("unable to check if the application has been initialised: %w", err),
		)

		return
	}

	if initialised {
		sendClientError(
			writer,
			http.StatusForbidden,
			ErrApplicationAlreadyInitialised,
		)

		return
	}

	switch request.Method {
	case http.MethodGet:
		s.getSetupForm(writer, request)

		return
	case http.MethodPost:
		s.setupAccount(writer, request)

		return
	default:
		sendClientError(
			writer,
			http.StatusMethodNotAllowed,
			fmt.Errorf("the setup endpoint does not support %s", request.Method),
		)

		return
	}
}

func (s *Server) getSetupForm(writer http.ResponseWriter, _ *http.Request) {
	form := formProfile{
		ProfileID: "",
		Password:  "",
		Profile: formProfileInfo{
			DisplayName: "",
			URL:         "",
			Email:       "",
			PhotoURL:    "",
		},
		FieldErrors: make(map[string]string),
	}

	generateAndSendHTMLResponse(
		writer,
		"setup",
		http.StatusOK,
		form,
	)
}

func (s *Server) setupAccount(writer http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		sendClientError(
			writer,
			http.StatusBadRequest,
			fmt.Errorf("error parsing the form: %w", err),
		)

		return
	}

	form := formProfile{
		ProfileID: request.PostFormValue("profileID"),
		Password:  request.PostFormValue("password"),
		Profile: formProfileInfo{
			DisplayName: request.PostFormValue("profileDisplayName"),
			URL:         request.PostFormValue("profileURL"),
			PhotoURL:    request.PostFormValue("profilePhotoURL"),
			Email:       request.PostFormValue("profileEmail"),
		},
		FieldErrors: make(map[string]string),
	}

	if validForm := form.validate(); !validForm {
		generateAndSendHTMLResponse(
			writer,
			"setup",
			http.StatusUnprocessableEntity,
			form,
		)

		return
	}

	hashedPassword, err := auth.HashPassword(form.Password)
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf("unable to hash the password: %w", err),
		)

		return
	}

	newProfile := database.Profile{
		HashedPassword: hashedPassword,
		Information: database.ProfileInformation{
			Name:     form.Profile.DisplayName,
			URL:      form.Profile.URL,
			PhotoURL: form.Profile.PhotoURL,
			Email:    form.Profile.Email,
		},
	}

	if err := database.CreateProfile(s.boltdb, form.ProfileID, newProfile); err != nil {
		sendServerError(
			writer,
			fmt.Errorf("unable to create the profile in the database: %w", err),
		)

		return
	}

	if err := database.UpdateInitialisedKey(s.boltdb); err != nil {
		sendServerError(
			writer,
			fmt.Errorf("unable to update the app's initialised flag: %w", err),
		)

		return
	}

	http.Redirect(writer, request, "/profile/login", http.StatusSeeOther)
}

type formProfile struct {
	ProfileID   string
	Password    string
	Profile     formProfileInfo
	FieldErrors map[string]string
}

type formProfileInfo struct {
	DisplayName string
	URL         string
	Email       string
	PhotoURL    string
}

func (f *formProfile) validate() bool {
	f.FieldErrors = make(map[string]string)

	canonicalisedWebsite, err := utilities.ValidateProfileURL(strings.TrimSpace(f.ProfileID))
	if err != nil {
		slog.Error("profile website validation failed", "error", err.Error())

		f.FieldErrors["ProfileID"] = "Please enter a valid website"
	} else {
		f.ProfileID = canonicalisedWebsite
	}

	minPasswordLength := 8
	if utf8.RuneCountInString(f.Password) < minPasswordLength {
		f.FieldErrors["Password"] = "The password must be at least 8 characters long"
	}

	return len(f.FieldErrors) == 0
}
