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
	ActiveTab        string
	ProfileID        string
	Title            string
	SettingsCategory string
	FailureMessage   string
	SuccessMessage   string
	FieldErrors      map[string]string
}

type settingsChangePasswordForm struct {
	currentPassword      string
	newPassword          string
	confirmedNewPassword string
}

func (f *settingsChangePasswordForm) valid() (map[string]string, bool) {
	fieldErrs := make(map[string]string)

	if utf8.RuneCountInString(f.newPassword) < 8 {
		fieldErrs["NewPassword"] = "The password must be at least 8 characters long"

		return fieldErrs, false
	}

	if f.newPassword != f.confirmedNewPassword {
		fieldErrs["NewPassword"] = "Your passwords did not match"
		fieldErrs["ConfirmedNewPassword"] = "Your passwords did not match"

		return fieldErrs, false
	}

	if f.newPassword == f.currentPassword {
		fieldErrs["NewPassword"] = "Your new password must be different from your current password"

		return fieldErrs, false
	}

	return nil, true
}

func (s *Server) getUpdatePasswordPage(writer http.ResponseWriter, _ *http.Request, profileID string) {
	page := settingsChangePasswordPage{
		ActiveTab:        activeTabSettings,
		ProfileID:        profileID,
		Title:            updateProfilePasswordPageTitle(),
		SettingsCategory: settingsPasswordChange,
		FailureMessage:   "",
		SuccessMessage:   "",
		FieldErrors:      map[string]string{},
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
	form := settingsChangePasswordForm{
		currentPassword:      request.PostFormValue("currentPassword"),
		newPassword:          request.PostFormValue("newPassword"),
		confirmedNewPassword: request.PostFormValue("confirmedNewPassword"),
	}

	fieldErrs, valid := form.valid()
	if !valid {
		page := settingsChangePasswordPage{
			ActiveTab:        activeTabSettings,
			ProfileID:        profileID,
			Title:            updateProfilePasswordPageTitle(),
			SettingsCategory: settingsPasswordChange,
			FailureMessage:   "Invalid form",
			SuccessMessage:   "",
			FieldErrors:      fieldErrs,
		}

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
		page := settingsChangePasswordPage{
			ActiveTab:        activeTabSettings,
			ProfileID:        profileID,
			Title:            updateProfilePasswordPageTitle(),
			SettingsCategory: settingsPasswordChange,
			FailureMessage:   "Unable to change the password",
			SuccessMessage:   "",
			FieldErrors:      map[string]string{},
		}

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

	err = auth.CheckPasswordHash(profile.HashedPassword, form.currentPassword)
	if err != nil {
		page := settingsChangePasswordPage{
			ActiveTab:        activeTabSettings,
			ProfileID:        profileID,
			Title:            updateProfilePasswordPageTitle(),
			SettingsCategory: settingsPasswordChange,
			FailureMessage:   "Unable to authorize the password change",
			SuccessMessage:   "",
			FieldErrors:      map[string]string{},
		}

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

	newHashedPassword, err := auth.HashPassword(form.newPassword)
	if err != nil {
		page := settingsChangePasswordPage{
			ActiveTab:        activeTabSettings,
			ProfileID:        profileID,
			Title:            updateProfilePasswordPageTitle(),
			SettingsCategory: settingsPasswordChange,
			FailureMessage:   "Unable to change the password",
			SuccessMessage:   "",
			FieldErrors:      map[string]string{},
		}

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
		page := settingsChangePasswordPage{
			ActiveTab:        activeTabSettings,
			ProfileID:        profileID,
			Title:            updateProfilePasswordPageTitle(),
			SettingsCategory: settingsPasswordChange,
			FailureMessage:   "Unable to change the password",
			SuccessMessage:   "",
			FieldErrors:      map[string]string{},
		}

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

	page := settingsChangePasswordPage{
		ActiveTab:        activeTabSettings,
		ProfileID:        profileID,
		Title:            updateProfilePasswordPageTitle(),
		SettingsCategory: settingsPasswordChange,
		FailureMessage:   "",
		SuccessMessage:   "Password changed successfully",
		FieldErrors:      map[string]string{},
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

func updateProfilePasswordPageTitle() string {
	return "Change password - Settings - " + info.ApplicationTitledName
}

func updateProfileInfoPageTitle() string {
	return "Update Profile information - Settings - " + info.ApplicationTitledName
}
