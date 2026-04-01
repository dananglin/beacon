// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"fmt"
	"net/http"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/info"
)

type overviewPage struct {
	ActiveTab   string
	ProfileID   string
	DisplayName string
	URL         string
	Email       string
	PhotoURL    string
	Title       string
}

func (s *Server) getOverviewPage(writer http.ResponseWriter, _ *http.Request, profileID string) {
	profileInfo, err := database.GetProfileInformation(s.boltdb, profileID)
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf("error getting the profile's information: %w", err),
		)

		return
	}

	page := overviewPage{
		ActiveTab:   activeTabHome,
		ProfileID:   profileID,
		DisplayName: profileInfo.Name,
		URL:         profileInfo.URL,
		Email:       profileInfo.Email,
		PhotoURL:    profileInfo.PhotoURL,
		Title:       "Your profile - " + info.ApplicationTitledName,
	}

	s.sendHTMLResponseWithTemplate(
		writer,
		"overview",
		http.StatusOK,
		page,
		nil,
		nil,
	)
}
