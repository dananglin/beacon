package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/config"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/info"
	bolt "go.etcd.io/bbolt"
)

type Server struct {
	httpServer *http.Server
	boltdb     *bolt.DB
	domainName string
	jwtSecret  string
}

func NewServer(configPath string) (*Server, error) {
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load the config: %w", err)
	}

	boltdb, err := database.New(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to open the database: %w", err)
	}

	server := Server{
		httpServer: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port),
			ReadHeaderTimeout: 1 * time.Second,
		},
		boltdb:     boltdb,
		domainName: cfg.Domain,
		jwtSecret:  cfg.JWT.Secret,
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

	mux.Handle("GET /{$}", http.HandlerFunc(s.rootRedirect))
	mux.Handle("GET /.well-known/oauth-authorization-server", setRequestID(http.HandlerFunc(s.getMetadata)))
	mux.Handle("GET /profile/login", setRequestID(http.HandlerFunc(s.getLoginForm)))
	mux.Handle("POST /profile/login", setRequestID(http.HandlerFunc(s.authenticate)))
	mux.Handle("GET /profile/login/confirmation", setRequestID(s.confirmation("Login successful.")))
	mux.Handle("GET /profile/setup", setRequestID(http.HandlerFunc(s.setup)))
	mux.Handle("POST /profile/setup", setRequestID(http.HandlerFunc(s.setup)))

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
		http.Redirect(writer, request, "/profile/setup", http.StatusSeeOther)

		return
	}

	http.Redirect(writer, request, "/profile/login", http.StatusSeeOther)
}
