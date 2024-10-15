package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/config"
	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/info"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg config.Config) *Server {
	address := fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port)

	httpServer := http.Server{
		Addr:              address,
		ReadHeaderTimeout: 1 * time.Second,
		Handler:           newMux(cfg),
	}

	server := Server{
		httpServer: &httpServer,
	}

	return &server
}

func (s *Server) Serve() error {
	slog.Info(info.ApplicationName+" is now ready to serve web requests", "address", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("error running the server: %w", err)
	}

	return nil
}
