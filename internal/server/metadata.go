package server

import (
	"fmt"
	"net/http"
)

func getMetadata(domain string) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		metadata := struct {
			Issuer                        string   `json:"issuer"`
			AuthorizationEndpoint         string   `json:"authorization_endpoint"`
			TokenEndpoint                 string   `json:"token_endpoint"`
			ServiceDocumentation          string   `json:"service_documentation"`
			CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`
		}{
			Issuer:                        fmt.Sprintf("https://%s/", domain),
			AuthorizationEndpoint:         fmt.Sprintf("https://%s/auth", domain),
			TokenEndpoint:                 fmt.Sprintf("https://%s/token", domain),
			ServiceDocumentation:          "https://indieauth.spec.indieweb.org",
			CodeChallengeMethodsSupported: []string{"S256"},
		}

		sendJSONResponse(writer, http.StatusOK, metadata)
	})
}
