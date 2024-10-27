// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"fmt"
	"net/http"
)

func (s *Server) getMetadata(writer http.ResponseWriter, _ *http.Request) {
	metadata := struct {
		Issuer                        string   `json:"issuer"`
		AuthorizationEndpoint         string   `json:"authorization_endpoint"`
		TokenEndpoint                 string   `json:"token_endpoint"`
		ServiceDocumentation          string   `json:"service_documentation"`
		CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`
	}{
		Issuer:                        fmt.Sprintf("https://%s/", s.domainName),
		AuthorizationEndpoint:         fmt.Sprintf("https://%s/auth", s.domainName),
		TokenEndpoint:                 fmt.Sprintf("https://%s/token", s.domainName),
		ServiceDocumentation:          "https://indieauth.spec.indieweb.org",
		CodeChallengeMethodsSupported: []string{"S256"},
	}

	sendJSONResponse(writer, http.StatusOK, metadata)
}
