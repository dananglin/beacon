// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import "net/http"

func redirectRoot(writer http.ResponseWriter, request *http.Request, _ string) {
	http.Redirect(writer, request, "/profile/overview", http.StatusSeeOther)
}

func (s *Server) redirectProfile(writer http.ResponseWriter, request *http.Request) {
	http.Redirect(writer, request, "/profile/overview", http.StatusSeeOther)
}

func (s *Server) redirectProfileSettings(writer http.ResponseWriter, request *http.Request) {
	http.Redirect(writer, request, "/profile/settings/info", http.StatusSeeOther)
}

func (s *Server) profileRedirectToLogin(writer http.ResponseWriter, request *http.Request) {
	redirectURL := "/profile/login?login_type=" + loginTypeProfile

	http.Redirect(writer, request, redirectURL, http.StatusSeeOther)
}
