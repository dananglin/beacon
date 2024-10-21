package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/config"
	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/database"
	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/info"
	bolt "go.etcd.io/bbolt"
)

type Server struct {
	httpServer *http.Server
	boltdb     *bolt.DB
	domainName string
}

func NewServer(configPath string) (*Server, error) {
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load the config: %w", err)
	}

	if cfg.Database.Path == "" {
		return nil, ErrMissingDatabasePath
	}

	boltdb, err := database.New(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to open the database: %w", err)
	}

	address := fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port)

	server := Server{
		httpServer: &http.Server{
			Addr:              address,
			ReadHeaderTimeout: 1 * time.Second,
		},
		boltdb:     boltdb,
		domainName: cfg.Domain,
	}

	server.setupRouter()

	return &server, nil
}

func (s *Server) Serve() error {
	slog.Info(info.ApplicationName+" is now ready to serve web requests", "address", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("error running the server: %w", err)
	}

	return nil
}

func (s *Server) setupRouter() {
	mux := http.NewServeMux()

	mux.Handle("GET /", http.HandlerFunc(s.rootRedirect))
	mux.Handle("GET /.well-known/oauth-authorization-server", setRequestID(http.HandlerFunc(s.getMetadata)))
	mux.Handle("GET /setup", setRequestID(http.HandlerFunc(s.setup)))
	mux.Handle("POST /setup", setRequestID(http.HandlerFunc(s.setup)))
	mux.Handle("GET /setup/confirmation", setRequestID(http.HandlerFunc(s.confirmation)))

	s.httpServer.Handler = mux
}

func (s *Server) rootRedirect(writer http.ResponseWriter, request *http.Request) {
	initialised, err := database.Initialised(s.boltdb)
	if err != nil {
		sendServerError(
			writer,
			fmt.Errorf("unable to check if the application has been initialised: %w", err),
		)

		return
	}

	if !initialised {
		http.Redirect(writer, request, "/setup", http.StatusSeeOther)
	}

	// TODO: redirect to the login page
}
