// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/auth"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/info"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
)

type setupPage struct {
	Title       string
	ProfileID   string
	DisplayName string
	Email       string
	URL         string
	PhotoURL    string
}

type setupForm struct {
	profileID         string
	password          string
	confirmedPassword string
	profile           struct {
		displayName string
		url         string
		email       string
		photoURL    string
	}
}

func (f *setupForm) validate() (fieldErrorLabel, error) {
	if strings.TrimSpace(f.profileID) == "" {
		return fieldErrorLabel{
			labelID: "profile_id_error",
			message: "This field must not be empty",
		}, formValidationError{reason: "the profile ID field is empty"}
	}

	if utf8.RuneCountInString(f.password) < 8 {
		return fieldErrorLabel{
			labelID: "password_error",
			message: "The password must be at least 8 characters long",
		}, formValidationError{reason: "the password field is less than 8 characters"}
	}

	if f.password != f.confirmedPassword {
		return fieldErrorLabel{
			labelID: "confirmed_password_error",
			message: "Your passwords do not match",
		}, formValidationError{reason: "the passwords do not match"}
	}

	return fieldErrorLabel{}, nil
}

func (s *Server) setup(writer http.ResponseWriter, request *http.Request) {
	dbInitialized, err := database.Initialized(s.boltdb)
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf("error checking if the database has been initialized or not: %w", err),
		)

		return
	}

	if dbInitialized {
		http.Redirect(writer, request, "/profile/login", http.StatusSeeOther)

		return
	}

	switch request.Method {
	case http.MethodGet:
		s.getSetupPage(writer, request)

		return
	case http.MethodPost:
		s.setupAccount(writer, request)

		return
	default:
		sendClientError(
			writer,
			http.StatusMethodNotAllowed,
			fmt.Errorf("the setup endpoint does not support the %q method", request.Method),
		)

		return
	}
}

func (s *Server) getSetupPage(writer http.ResponseWriter, _ *http.Request) {
	page := setupPage{
		Title:       setupPageTitle(),
		ProfileID:   "",
		DisplayName: "",
		Email:       "",
		URL:         "",
		PhotoURL:    "",
	}

	s.sendHTMLResponseWithTemplate(
		writer,
		"setup",
		http.StatusOK,
		page,
		nil,
		nil,
	)
}

func (s *Server) setupAccount(writer http.ResponseWriter, request *http.Request) {
	form := setupForm{
		profileID:         request.PostFormValue("profileID"),
		password:          request.PostFormValue("password"),
		confirmedPassword: request.PostFormValue("confirmedPassword"),
		profile: struct {
			displayName string
			url         string
			email       string
			photoURL    string
		}{
			displayName: request.PostFormValue("profileDisplayName"),
			url:         request.PostFormValue("profileURL"),
			photoURL:    request.PostFormValue("profilePhotoURL"),
			email:       request.PostFormValue("profileEmail"),
		},
	}

	fieldErrorLabel, err := form.validate()
	if err != nil {
		writer.Header().Set("HX-Retarget", "#"+fieldErrorLabel.labelID)
		writer.Header().Set("HX-Reswap", "outerHTML")

		s.sendHTMLResponse(
			writer,
			fmt.Appendf(
				[]byte{},
				responselabelErrorFmt,
				fieldErrorLabel.labelID,
				fieldErrorLabel.message,
			),
			http.StatusUnprocessableEntity,
			fmt.Errorf("error validating the form: %w", err),
			nil,
		)

		return
	}

	canonicalisedProfielID, err := utilities.ValidateAndCanonicalizeURL(
		strings.TrimSpace(form.profileID),
		false,
	)
	if err != nil {
		writer.Header().Set("HX-Retarget", "#profile_id_error")
		writer.Header().Set("HX-Reswap", "outerHTML")

		s.sendHTMLResponse(
			writer,
			fmt.Appendf(
				[]byte{},
				responselabelErrorFmt,
				"profile_id_error",
				"Please enter a valid domain or website",
			),
			http.StatusUnprocessableEntity,
			fmt.Errorf("error validating the profile ID: %w", err),
			nil,
		)

		return
	}

	// Hash the password
	hashedPassword, err := auth.HashPassword(form.password)
	if err != nil {
		s.sendHTMLResponse(
			writer,
			fmt.Appendf([]byte{}, responseFailureFmt, "Unable to set up your profile"),
			http.StatusInternalServerError,
			nil,
			fmt.Errorf("error hashing the password: %w", err),
		)

		return
	}

	// Create and store the new profile
	profile := database.Profile{
		HashedPassword: hashedPassword,
		Information: database.ProfileInformation{
			Name:     form.profile.displayName,
			URL:      form.profile.url,
			PhotoURL: form.profile.photoURL,
			Email:    form.profile.email,
		},
	}

	if err := database.Setup(s.boltdb, canonicalisedProfielID, profile); err != nil {
		s.sendHTMLResponse(
			writer,
			fmt.Appendf([]byte{}, responseFailureFmt, "Unable to set up your profile"),
			http.StatusInternalServerError,
			nil,
			fmt.Errorf("error setting up the database: %w", err),
		)

		return
	}

	// Ensure that the database has been initialized successfully.
	dbInitialized, err := database.Initialized(s.boltdb)
	if err != nil {
		s.sendHTMLResponse(
			writer,
			fmt.Appendf([]byte{}, responseFailureFmt, "Unable to set up your profile"),
			http.StatusInternalServerError,
			nil,
			fmt.Errorf("error checking if the database has been initialized or not: %w", err),
		)

		return
	}

	if !dbInitialized {
		s.sendHTMLResponse(
			writer,
			fmt.Appendf([]byte{}, responseFailureFmt, "Unable to set up your profile"),
			http.StatusInternalServerError,
			nil,
			ErrDatabaseNotInitialized,
		)

		return
	}

	s.dbInitialized = dbInitialized

	writer.Header().Set("Hx-Redirect", "/profile/login")
}

func setupPageTitle() string {
	return "Setup - " + info.ApplicationTitledName
}
