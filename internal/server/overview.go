// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"fmt"
	"net/http"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
)

type overviewPage struct {
	ProfileID   string
	DisplayName string
	URL         string
	Email       string
	PhotoURL    string
}

func (s *Server) getOverviewPage(writer http.ResponseWriter, _ *http.Request, profileID string) {
	info, err := database.GetProfileInformation(s.boltdb, profileID)
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf("error getting the profile's information: %w", err),
		)

		return
	}

	page := overviewPage{
		ProfileID:   profileID,
		DisplayName: info.Name,
		URL:         info.URL,
		Email:       info.Email,
		PhotoURL:    info.PhotoURL,
	}

	generateAndSendHTMLResponse(
		writer,
		"overview",
		http.StatusOK,
		page,
	)
}

func (s *Server) updateProfileInformation(writer http.ResponseWriter, request *http.Request, profileID string) {
	if err := request.ParseForm(); err != nil {
		sendClientError(
			writer,
			http.StatusBadRequest,
			fmt.Errorf("error parsing the form: %w", err),
		)

		return
	}

	newProfileInfo := database.ProfileInformation{
		Name:     request.PostFormValue("profileDisplayName"),
		URL:      request.PostFormValue("profileURL"),
		PhotoURL: request.PostFormValue("profilePhotoURL"),
		Email:    request.PostFormValue("profileEmail"),
	}

	if err := database.UpdateProfileInformation(s.boltdb, profileID, newProfileInfo); err != nil {
		sendServerError(
			writer,
			fmt.Errorf("error updating profile: %w", err),
		)

		return
	}

	http.Redirect(writer, request, "/profile/overview", http.StatusSeeOther)
}
