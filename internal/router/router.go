package router

import "net/http"

func NewServeMux() *http.ServeMux {
	mux := http.NewServeMux()

	return mux
}
