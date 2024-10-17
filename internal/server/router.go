package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"

	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/config"
)

func newMux(cfg config.Config) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /.well-known/oauth-authorization-server", setRequestID(metadataHandler(cfg.Domain)))

	return mux
}

func setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestID := "UNKNOWN"
		id := make([]byte, 16)

		if _, err := rand.Read(id); err != nil {
			slog.Error("unable to create the request ID.", "error", err.Error())
		} else {
			requestID = hex.EncodeToString(id)
		}

		writer.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(writer, request)
	})
}

func metadataHandler(domain string) http.Handler {
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

		sendResponse(writer, http.StatusOK, metadata)
	})
}
