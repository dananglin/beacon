package server

import (
	"fmt"
	"net/http"

	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/config"
)

func newMux(cfg config.Config) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /.well-known/oauth-authorization-server", metadataHandleFunc(cfg.Domain))

	return mux
}

func metadataHandleFunc(domain string) http.HandlerFunc {
	return func(writer http.ResponseWriter, _ *http.Request) {
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
	}
}
