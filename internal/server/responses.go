package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func sendJSONResponse(w http.ResponseWriter, statusCode int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Error marshalling the response to JSON", "error", err.Error())

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(data)
}
