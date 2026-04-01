// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
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
}

func (s *Server) getUpdateProfileInfoPage(writer http.ResponseWriter, _ *http.Request, profileID string) {
	profileInfo, err := database.GetProfileInformation(s.boltdb, profileID)
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
	}

	s.sendHTMLResponseWithTemplate(
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
		s.sendHTMLResponse(
			writer,
			fmt.Appendf([]byte{}, responseFailureFmt, "Unable to update your profile"),
			http.StatusInternalServerError,
			nil,
			fmt.Errorf("error updating the profile: %w", err),
		)

		return
	}

	s.sendHTMLResponse(
		writer,
		fmt.Appendf([]byte{}, responseSuccessFmt, "Successfully updated your profile information"),
		http.StatusOK,
		nil,
		nil,
	)
}

type settingsChangePasswordPage struct {
	ActiveTab        string
	ProfileID        string
	Title            string
	SettingsCategory string
}

type settingsChangePasswordForm struct {
	currentPassword      string
	newPassword          string
	confirmedNewPassword string
}

func (f *settingsChangePasswordForm) validate() (fieldErrorLabel, error) {
	if utf8.RuneCountInString(f.newPassword) < 8 {
		return fieldErrorLabel{
			labelID: "new_password_error",
			message: "The new password must be at least 8 characters long",
		}, formValidationError{reason: "the password is less than 8 characters long"}
	}

	if f.newPassword != f.confirmedNewPassword {
		return fieldErrorLabel{
			labelID: "confirmed_password_error",
			message: "Your passwords do not match",
		}, formValidationError{"the #new_password and #confirmed_password values do not match"}
	}

	if f.newPassword == f.currentPassword {
		return fieldErrorLabel{
			labelID: "new_password_error",
			message: "The new password must be different from your current password",
		}, formValidationError{"the new password is the same as the current password"}
	}

	return fieldErrorLabel{}, nil
}

func (s *Server) getUpdatePasswordPage(writer http.ResponseWriter, _ *http.Request, profileID string) {
	page := settingsChangePasswordPage{
		ActiveTab:        activeTabSettings,
		ProfileID:        profileID,
		Title:            updateProfilePasswordPageTitle(),
		SettingsCategory: settingsPasswordChange,
	}

	s.sendHTMLResponseWithTemplate(
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

	profile, err := database.GetProfile(s.boltdb, profileID)
	if err != nil {
		s.sendHTMLResponse(
			writer,
			fmt.Appendf([]byte{}, responseFailureFmt, "Unable to change the password"),
			http.StatusInternalServerError,
			nil,
			fmt.Errorf("error retrieving the profile: %w", err),
		)

		return
	}

	err = auth.CheckPasswordHash(profile.HashedPassword, form.currentPassword)
	if err != nil {
		s.sendHTMLResponse(
			writer,
			fmt.Appendf([]byte{}, responseFailureFmt, "Unable to authorize the password change"),
			http.StatusUnauthorized,
			fmt.Errorf("password verification failed: %w", err),
			nil,
		)

		return
	}

	newHashedPassword, err := auth.HashPassword(form.newPassword)
	if err != nil {
		s.sendHTMLResponse(
			writer,
			fmt.Appendf([]byte{}, responseFailureFmt, "Unable to change the password"),
			http.StatusInternalServerError,
			nil,
			fmt.Errorf("error hashing the password: %w", err),
		)

		return
	}

	err = database.UpdateHashedPassword(s.boltdb, profileID, newHashedPassword)
	if err != nil {
		s.sendHTMLResponse(
			writer,
			fmt.Appendf([]byte{}, responseFailureFmt, "Unable to change the password"),
			http.StatusInternalServerError,
			nil,
			fmt.Errorf("error saving the new password to the database: %w", err),
		)

		return
	}

	s.sendHTMLResponse(
		writer,
		fmt.Appendf([]byte{}, responseSuccessFmt, "Password changed successfully"),
		http.StatusOK,
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
