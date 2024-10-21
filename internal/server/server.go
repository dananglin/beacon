package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/config"
	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/database"
	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/info"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(configPath string) (*Server, error) {
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load the config: %w", err)
	}

	if cfg.Database.Path == "" {
		return nil, fmt.Errorf("please set the database path")
	}

	boltdb, err := database.New(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to open the database: %w", err)
	}

	address := fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port)

	httpServer := http.Server{
		Addr:              address,
		ReadHeaderTimeout: 1 * time.Second,
		Handler:           newMux(cfg, boltdb),
	}

	server := Server{
		httpServer: &httpServer,
	}

	return &server, nil
}

func (s *Server) Serve() error {
	slog.Info(info.ApplicationName+" is now ready to serve web requests", "address", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("error running the server: %w", err)
	}

	return nil
}
