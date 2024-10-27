// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
)

func sendJSONResponse(writer http.ResponseWriter, statusCode int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf("error marshalling the JSON response: %w", err),
		)

		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	_, _ = writer.Write(data)
}

func generateAndSendHTMLResponse(writer http.ResponseWriter, templateName string, statusCode int, data any) {
	tmpl, err := template.New(templateName).ParseFS(templates, "templates/*")
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf("error creating the HTML template: %w", err),
		)

		return
	}

	writer.Header().Set("Content-Type", "text/html")
	writer.WriteHeader(statusCode)

	if err := tmpl.Execute(writer, data); err != nil {
		sendServerError(
			writer,
			fmt.Errorf("error generating HTML from the template: %w", err),
		)

		return
	}
}

func sendClientError(writer http.ResponseWriter, statusCode int, err error) {
	sendErrorResponse(
		writer,
		statusCode,
		"Client Error: "+err.Error(),
	)
}

func sendServerError(writer http.ResponseWriter, err error) {
	sendErrorResponse(
		writer,
		http.StatusInternalServerError,
		"Server Error: "+err.Error(),
	)
}

func sendErrorResponse(writer http.ResponseWriter, statusCode int, message string) {
	slog.Error(message, "request-id", writer.Header().Get("X-Request-ID"))

	http.Error(writer, http.StatusText(statusCode), statusCode)
}
