// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"net/http"
)

type responseWriter struct {
	http.ResponseWriter

	headerWritten bool
	statusCode    int
	wrote         int
}

func newResponseWriter(writer http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: writer,
		headerWritten:  false,
		statusCode:     http.StatusOK,
		wrote:          0,
	}
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)

	if !w.headerWritten {
		w.statusCode = statusCode
		w.headerWritten = true
	}
}

func (w *responseWriter) Write(bytes []byte) (int, error) {
	w.headerWritten = true
	wrote, err := w.ResponseWriter.Write(bytes)
	w.wrote = wrote
	return wrote, err
}

func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
