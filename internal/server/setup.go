// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"errors"
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
	Title          string
	FailureMessage string
	ProfileID      string
	DisplayName    string
	Email          string
	URL            string
	PhotoURL       string
	FieldErrors    map[string]string
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

func (f *setupForm) valid() (map[string]string, bool) {
	fieldErrors := make(map[string]string)

	if strings.TrimSpace(f.profileID) == "" {
		fieldErrors["ProfileID"] = "This field must not be empty"

		return fieldErrors, false
	}

	if utf8.RuneCountInString(f.password) < 8 {
		fieldErrors["Password"] = "The password must be at least 8 characters long"

		return fieldErrors, false
	}

	if f.password != f.confirmedPassword {
		fieldErrors["Password"] = "Your passwords do not match"
		fieldErrors["ConfirmedPassword"] = "Your passwords do not match"

		return fieldErrors, false
	}

	return nil, true
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
		Title:          setupPageTitle(),
		FailureMessage: "",
		ProfileID:      "",
		DisplayName:    "",
		Email:          "",
		URL:            "",
		PhotoURL:       "",
		FieldErrors:    make(map[string]string),
	}

	s.sendHTMLResponse(
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

	fieldErrs, valid := form.valid()
	if !valid {
		page := setupPage{
			Title:          setupPageTitle(),
			FailureMessage: "Invalid form",
			ProfileID:      form.profileID,
			DisplayName:    form.profile.displayName,
			Email:          form.profile.email,
			URL:            form.profile.url,
			PhotoURL:       form.profile.photoURL,
			FieldErrors:    fieldErrs,
		}

		s.sendHTMLResponse(
			writer,
			"setup",
			http.StatusUnprocessableEntity,
			page,
			errors.New("invalid form"),
			nil,
		)

		return
	}

	canonicalisedProfielID, err := utilities.ValidateAndCanonicalizeURL(
		strings.TrimSpace(form.profileID),
		false,
	)
	if err != nil {
		page := setupPage{
			Title:          setupPageTitle(),
			FailureMessage: "Invalid profile ID",
			ProfileID:      form.profileID,
			DisplayName:    form.profile.displayName,
			Email:          form.profile.email,
			URL:            form.profile.url,
			PhotoURL:       form.profile.photoURL,
			FieldErrors: map[string]string{
				"ProfileID": "Please enter a valid domain or website",
			},
		}

		s.sendHTMLResponse(
			writer,
			"setup",
			http.StatusUnprocessableEntity,
			page,
			fmt.Errorf("error validating the profile ID: %w", err),
			nil,
		)

		return
	}

	hashedPassword, err := auth.HashPassword(form.password)
	if err != nil {
		page := setupPage{
			Title:          setupPageTitle(),
			FailureMessage: "Invalid form",
			ProfileID:      form.profileID,
			DisplayName:    form.profile.displayName,
			Email:          form.profile.email,
			URL:            form.profile.url,
			PhotoURL:       form.profile.photoURL,
			FieldErrors:    make(map[string]string),
		}

		s.sendHTMLResponse(
			writer,
			"setup",
			http.StatusInternalServerError,
			page,
			nil,
			fmt.Errorf("error hashing the password: %w", err),
		)

		return
	}

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
		page := setupPage{
			Title:          setupPageTitle(),
			FailureMessage: "Unable to create your profile",
			ProfileID:      form.profileID,
			DisplayName:    form.profile.displayName,
			Email:          form.profile.email,
			URL:            form.profile.url,
			PhotoURL:       form.profile.photoURL,
			FieldErrors:    make(map[string]string),
		}

		s.sendHTMLResponse(
			writer,
			"setup",
			http.StatusInternalServerError,
			page,
			nil,
			fmt.Errorf("error setting up the database: %w", err),
		)

		return
	}

	// Ensure that the database has been initialized successfully.
	dbInitialized, err := database.Initialized(s.boltdb)
	if err != nil {
		page := setupPage{
			Title:          setupPageTitle(),
			FailureMessage: "Unable to create your profile",
			ProfileID:      form.profileID,
			DisplayName:    form.profile.displayName,
			Email:          form.profile.email,
			URL:            form.profile.url,
			PhotoURL:       form.profile.photoURL,
			FieldErrors:    make(map[string]string),
		}

		s.sendHTMLResponse(
			writer,
			"setup",
			http.StatusInternalServerError,
			page,
			nil,
			fmt.Errorf("error checking if the database has been initialized or not: %w", err),
		)

		return
	}

	if !dbInitialized {
		page := setupPage{
			Title:          setupPageTitle(),
			FailureMessage: "Unable to create your profile",
			ProfileID:      form.profileID,
			DisplayName:    form.profile.displayName,
			Email:          form.profile.email,
			URL:            form.profile.url,
			PhotoURL:       form.profile.photoURL,
			FieldErrors:    make(map[string]string),
		}

		s.sendHTMLResponse(
			writer,
			"setup",
			http.StatusInternalServerError,
			page,
			nil,
			ErrDatabaseNotInitialized,
		)

		return
	}

	s.dbInitialized = dbInitialized

	http.Redirect(writer, request, "/profile/login", http.StatusSeeOther)
}

func setupPageTitle() string {
	return "Setup - " + info.ApplicationTitledName
}
