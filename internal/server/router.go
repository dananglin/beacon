package server

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"

	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/config"
	bolt "go.etcd.io/bbolt"
)

func newMux(cfg config.Config, boltdb *bolt.DB) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /.well-known/oauth-authorization-server", setRequestID(getMetadata(cfg.Domain)))
	mux.Handle("GET /setup", setRequestID(getSetupForm(boltdb)))
	mux.Handle("POST /setup", setRequestID(setupAccount(boltdb)))
	mux.Handle("GET /setup/confirmation", setRequestID(http.HandlerFunc(confirmation)))

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
