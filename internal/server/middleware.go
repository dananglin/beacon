package server

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
)

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
