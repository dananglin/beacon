// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func (s *Server) sendHTMLResponseWithTemplate(
	writer http.ResponseWriter,
	templateName string,
	statusCode int,
	data any,
	clientErr error,
	serverErr error,
) {
	if clientErr != nil {
		slog.LogAttrs(
			context.Background(),
			slog.LevelError,
			"Client error",
			slog.Any("error", clientErr),
			slog.String("request_id", writer.Header().Get("X-Request-ID")),
		)
	}

	if serverErr != nil {
		slog.LogAttrs(
			context.Background(),
			slog.LevelError,
			"Server error",
			slog.Any("error", serverErr),
			slog.String("request_id", writer.Header().Get("X-Request-ID")),
		)
	}

	writer.Header().Set("Content-Type", "text/html; charset=UTF-8")
	writer.WriteHeader(statusCode)

	if err := s.htmlTemplate.ExecuteTemplate(writer, templateName, data); err != nil {
		sendServerError(
			writer,
			fmt.Errorf("error generating HTML from the template: %w", err),
		)

		return
	}
}

func (s *Server) sendHTMLResponse(
	writer http.ResponseWriter,
	payload []byte,
	statusCode int,
	clientErr error,
	serverErr error,
) {
	if clientErr != nil {
		slog.LogAttrs(
			context.Background(),
			slog.LevelError,
			"Client error",
			slog.Any("error", clientErr),
			slog.String("request_id", writer.Header().Get("X-Request-ID")),
		)
	}

	if serverErr != nil {
		slog.LogAttrs(
			context.Background(),
			slog.LevelError,
			"Server error",
			slog.Any("error", serverErr),
			slog.String("request_id", writer.Header().Get("X-Request-ID")),
		)
	}

	writer.Header().Set("Content-Type", "text/html; charset=UTF-8")
	writer.WriteHeader(statusCode)
	_, _ = writer.Write(payload)
}

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

func sendClientError(writer http.ResponseWriter, statusCode int, err error) {
	sendErrorResponse(
		writer,
		statusCode,
		"Client error",
		err,
	)
}

func sendServerError(writer http.ResponseWriter, err error) {
	sendErrorResponse(
		writer,
		http.StatusInternalServerError,
		"Server error",
		err,
	)
}

func sendErrorResponse(writer http.ResponseWriter, statusCode int, msg string, err error) {
	slog.LogAttrs(
		context.Background(),
		slog.LevelError,
		msg,
		slog.Any("error", err),
		slog.String("request_id", writer.Header().Get("X-Request-ID")),
	)

	http.Error(writer, http.StatusText(statusCode), statusCode)
}
