// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"errors"
	"fmt"
	"net/http"
	"unicode/utf8"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/auth"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/info"
)

const (
	settingsProfileInfo    = "profile_info"
	settingsPasswordChange = "password_change"
)

type settingsUpdateProfileInfoPage struct {
	ActiveTab        string
	ProfileID        string
	DisplayName      string
	URL              string
	Email            string
	PhotoURL         string
	Title            string
	SettingsCategory string
	FailureMessage   string
	SuccessMessage   string
}

func (s *Server) getUpdateProfileInfoPage(writer http.ResponseWriter, _ *http.Request, profileID string) {
	profileInfo, err := database.GetProfileInformation(
		s.boltdb,
		profileID,
	)
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf("error getting the profile's information: %w", err),
		)

		return
	}

	page := settingsUpdateProfileInfoPage{
		ActiveTab:        activeTabSettings,
		ProfileID:        profileID,
		DisplayName:      profileInfo.Name,
		URL:              profileInfo.URL,
		Email:            profileInfo.Email,
		PhotoURL:         profileInfo.PhotoURL,
		Title:            updateProfileInfoPageTitle(),
		SettingsCategory: settingsProfileInfo,
		FailureMessage:   "",
		SuccessMessage:   "",
	}

	s.sendHTMLResponse(
		writer,
		"settings",
		http.StatusOK,
		page,
		nil,
		nil,
	)
}

func (s *Server) updateProfileInformation(writer http.ResponseWriter, request *http.Request, profileID string) {
	newProfileInfo := database.ProfileInformation{
		Name:     request.PostFormValue("profileDisplayName"),
		URL:      request.PostFormValue("profileURL"),
		PhotoURL: request.PostFormValue("profilePhotoURL"),
		Email:    request.PostFormValue("profileEmail"),
	}

	if err := database.UpdateProfileInformation(s.boltdb, profileID, newProfileInfo); err != nil {
		page := settingsUpdateProfileInfoPage{
			ActiveTab:        activeTabSettings,
			ProfileID:        profileID,
			DisplayName:      newProfileInfo.Name,
			URL:              newProfileInfo.URL,
			Email:            newProfileInfo.Email,
			PhotoURL:         newProfileInfo.PhotoURL,
			Title:            updateProfileInfoPageTitle(),
			SettingsCategory: settingsProfileInfo,
			FailureMessage:   "Unable to update your profile",
			SuccessMessage:   "",
		}

		s.sendHTMLResponse(
			writer,
			"settings",
			http.StatusInternalServerError,
			page,
			nil,
			fmt.Errorf("error updating the profile: %w", err),
		)

		return
	}

	page := settingsUpdateProfileInfoPage{
		ActiveTab:        activeTabSettings,
		ProfileID:        profileID,
		DisplayName:      newProfileInfo.Name,
		URL:              newProfileInfo.URL,
		Email:            newProfileInfo.Email,
		PhotoURL:         newProfileInfo.PhotoURL,
		Title:            updateProfileInfoPageTitle(),
		SettingsCategory: settingsProfileInfo,
		FailureMessage:   "",
		SuccessMessage:   "Successfully updated your profile information",
	}

	s.sendHTMLResponse(
		writer,
		"settings",
		http.StatusOK,
		page,
		nil,
		nil,
	)
}

type settingsChangePasswordPage struct {
	ActiveTab            string
	ProfileID            string
	CurrentPassword      string
	NewPassword          string
	ConfirmedNewPassword string
	Title                string
	SettingsCategory     string
	FailureMessage       string
	SuccessMessage       string
	FieldErrors          map[string]string
}

func (p *settingsChangePasswordPage) valid() bool {
	if utf8.RuneCountInString(p.NewPassword) < 8 {
		p.FieldErrors["NewPassword"] = "The password must be at least 8 characters long"

		return false
	}

	if p.NewPassword != p.ConfirmedNewPassword {
		p.FieldErrors["NewPassword"] = "Your passwords did not match"
		p.FieldErrors["ConfirmedNewPassword"] = "Your passwords did not match"

		return false
	}

	if p.NewPassword == p.CurrentPassword {
		p.FieldErrors["NewPassword"] = "Your new password must be different from your current password"

		return false
	}

	return true
}

func (p *settingsChangePasswordPage) clearFields() {
	p.CurrentPassword = ""
	p.NewPassword = ""
	p.ConfirmedNewPassword = ""
}

func (s *Server) getUpdatePasswordPage(writer http.ResponseWriter, _ *http.Request, profileID string) {
	page := settingsChangePasswordPage{
		ActiveTab:            activeTabSettings,
		ProfileID:            profileID,
		CurrentPassword:      "",
		NewPassword:          "",
		ConfirmedNewPassword: "",
		FailureMessage:       "",
		SuccessMessage:       "",
		Title:                updateProfilePasswordPageTitle(),
		SettingsCategory:     settingsPasswordChange,
		FieldErrors:          make(map[string]string),
	}

	s.sendHTMLResponse(
		writer,
		"settings",
		http.StatusOK,
		page,
		nil,
		nil,
	)
}

func (s *Server) updateProfilePassword(writer http.ResponseWriter, request *http.Request, profileID string) {
	page := settingsChangePasswordPage{
		ActiveTab:            activeTabSettings,
		ProfileID:            profileID,
		CurrentPassword:      request.PostFormValue("currentPassword"),
		NewPassword:          request.PostFormValue("newPassword"),
		ConfirmedNewPassword: request.PostFormValue("confirmedNewPassword"),
		Title:                updateProfilePasswordPageTitle(),
		SettingsCategory:     settingsPasswordChange,
		FieldErrors:          make(map[string]string),
	}

	if valid := page.valid(); !valid {
		page.clearFields()
		page.FailureMessage = "Invalid form"

		s.sendHTMLResponse(
			writer,
			"settings",
			http.StatusUnprocessableEntity,
			page,
			errors.New("invalid form"),
			nil,
		)

		return
	}

	profile, err := database.GetProfile(s.boltdb, profileID)
	if err != nil {
		page.clearFields()
		page.FailureMessage = "Unable change the password"

		s.sendHTMLResponse(
			writer,
			"settings",
			http.StatusInternalServerError,
			page,
			nil,
			fmt.Errorf("error retrieving the profile: %w", err),
		)

		return
	}

	err = auth.CheckPasswordHash(profile.HashedPassword, page.CurrentPassword)
	if err != nil {
		page.clearFields()
		page.FailureMessage = "You are not authorized to change the password"

		s.sendHTMLResponse(
			writer,
			"settings",
			http.StatusUnauthorized,
			page,
			fmt.Errorf("password verification failed: %w", err),
			nil,
		)

		return
	}

	newHashedPassword, err := auth.HashPassword(page.NewPassword)
	if err != nil {
		page.clearFields()
		page.FailureMessage = "Unable to change the password"

		s.sendHTMLResponse(
			writer,
			"settings",
			http.StatusInternalServerError,
			page,
			nil,
			fmt.Errorf("error hashing the password: %w", err),
		)

		return
	}

	err = database.UpdateHashedPassword(s.boltdb, profileID, newHashedPassword)
	if err != nil {
		page.clearFields()
		page.FailureMessage = "Unable to change the password"
		s.sendHTMLResponse(
			writer,
			"settings",
			http.StatusInternalServerError,
			page,
			nil,
			fmt.Errorf("error saving the new password to the database: %w", err),
		)

		return
	}

	page.clearFields()
	page.SuccessMessage = "Password changed successfully"

	s.sendHTMLResponse(
		writer,
		"settings",
		http.StatusOK,
		page,
		nil,
		nil,
	)
}

func updateProfilePasswordPageTitle() string {
	return "Change password - Settings - " + info.ApplicationTitledName
}

func updateProfileInfoPageTitle() string {
	return "Update Profile information - Settings - " + info.ApplicationTitledName
}
