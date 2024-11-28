// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"net/http"
)

type metadata struct {
	Issuer                                 string   `json:"issuer"`
	AuthorizationEndpoint                  string   `json:"authorization_endpoint"`
	TokenEndpoint                          string   `json:"token_endpoint"`
	ServiceDocumentation                   string   `json:"service_documentation"`
	CodeChallengeMethodsSupported          []string `json:"code_challenge_methods_supported"`
	GrantTypesSupported                    []string `json:"grant_types_supported"`
	ResponseTypesSupported                 []string `json:"response_types_supported"`
	ScopesSupported                        []string `json:"scopes_supported"`
	AuthorizationResponseISSParamSupported bool     `json:"authorization_response_iss_parameter_supported"`
}

func (s *Server) getMetadata(writer http.ResponseWriter, _ *http.Request) {
	metadata := metadata{
		Issuer:                                 s.issuer,
		AuthorizationEndpoint:                  s.authEndpoint,
		TokenEndpoint:                          s.tokenEndpoint,
		ServiceDocumentation:                   "https://indieauth.spec.indieweb.org",
		CodeChallengeMethodsSupported:          []string{"S256"},
		GrantTypesSupported:                    []string{"authorization_code"},
		ResponseTypesSupported:                 []string{"code"},
		ScopesSupported:                        []string{"profile", "email"},
		AuthorizationResponseISSParamSupported: true,
	}

	sendJSONResponse(writer, http.StatusOK, metadata)
}
