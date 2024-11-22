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
	dbInitialized, err := database.Initialized(s.boltdb)
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf("unable to check if the database has been initialized or not: %w", err),
		)

		return
	}

	if dbInitialized {
		sendClientError(
			writer,
			http.StatusForbidden,
			ErrDatabaseAlreadyInitialized,
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
	form := setupForm{
		ProfileID:         "",
		Password:          "",
		ConfirmedPassword: "",
		Profile: setupFormProfile{
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

	form := setupForm{
		ProfileID:         request.PostFormValue("profileID"),
		Password:          request.PostFormValue("password"),
		ConfirmedPassword: request.PostFormValue("confirmedPassword"),
		Profile: setupFormProfile{
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

	profile := database.Profile{
		HashedPassword: hashedPassword,
		Information: database.ProfileInformation{
			Name:     form.Profile.DisplayName,
			URL:      form.Profile.URL,
			PhotoURL: form.Profile.PhotoURL,
			Email:    form.Profile.Email,
		},
	}

	if err := database.Setup(s.boltdb, form.ProfileID, profile); err != nil {
		sendServerError(
			writer,
			fmt.Errorf("error setting up the database: %w", err),
		)

		return
	}

	// Ensure that the database has been initialized successfully.
	dbInitialized, err := database.Initialized(s.boltdb)
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf("unable to check if the database has been initialized or not: %w", err),
		)

		return
	}

	if !dbInitialized {
		sendServerError(
			writer,
			ErrDatabaseNotInitialized,
		)

		return
	}

	s.dbInitialized = dbInitialized

	http.Redirect(writer, request, "/profile/login", http.StatusSeeOther)
}

type setupForm struct {
	ProfileID         string
	Password          string
	ConfirmedPassword string
	Profile           setupFormProfile
	FieldErrors       map[string]string
}

type setupFormProfile struct {
	DisplayName string
	URL         string
	Email       string
	PhotoURL    string
}

func (f *setupForm) validate() bool {
	f.FieldErrors = make(map[string]string)

	canonicalisedProfielID, err := utilities.ValidateAndCanonicalizeURL(strings.TrimSpace(f.ProfileID))
	if err != nil {
		slog.Error("profile ID validation failed", "error", err.Error())

		f.FieldErrors["ProfileID"] = "Please enter a valid domain or website"
	} else {
		f.ProfileID = canonicalisedProfielID
	}

	minPasswordLength := 8
	if utf8.RuneCountInString(f.Password) < minPasswordLength {
		f.FieldErrors["Password"] = "The password must be at least 8 characters long"
	}

	if f.Password != f.ConfirmedPassword {
		f.FieldErrors["ConfirmedPassword"] = "Your passwords do not match"
	}

	return len(f.FieldErrors) == 0
}
