package server

import "net/http"

func (s *Server) confirmation(message string) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		generateAndSendHTMLResponse(
			writer,
			"confirmation",
			http.StatusOK,
			message,
		)
	})
}
