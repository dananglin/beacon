// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/cache"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/config"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/info"
	bolt "go.etcd.io/bbolt"
)

const (
	templatesFSDir = "ui/templates"
	staticFSDir    = "ui/static"
)

//go:embed ui/templates/*
var templatesFS embed.FS

//go:embed ui/static/*
var staticFS embed.FS

type (
	profileHandlerFunc  func(writer http.ResponseWriter, request *http.Request, profileID string)
	exchangeHandlerFunc func(writer http.ResponseWriter, data clientRequestData)

	Server struct {
		httpServer              *http.Server
		boltdb                  *bolt.DB
		cache                   *cache.Cache
		dbInitialized           bool
		domainName              string
		jwtSecret               string
		jwtCookieName           string
		authPath                string
		authEndpoint            string
		issuer                  string
		tokenEndpoint           string
		tokenPath               string
		gracefulShutdownTimeout time.Duration
	}
)

func NewServer(configPath string) (*Server, error) {
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load the config: %w", err)
	}

	boltdb, err := database.Open(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to open the database: %w", err)
	}

	authPath := "/indieauth/authorize"
	tokenPath := "/indieauth/token" // #nosec G101 -- This is not hardcoded credentials.

	server := Server{
		httpServer: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port),
			ReadHeaderTimeout: 1 * time.Second,
		},
		boltdb:                  boltdb,
		cache:                   cache.NewCache(1 * time.Minute),
		domainName:              cfg.Domain,
		jwtSecret:               cfg.JWT.Secret,
		jwtCookieName:           cfg.JWT.CookieName,
		authPath:                authPath,
		authEndpoint:            fmt.Sprintf("https://%s%s", cfg.Domain, authPath),
		issuer:                  fmt.Sprintf("https://%s/", cfg.Domain),
		tokenPath:               tokenPath,
		tokenEndpoint:           fmt.Sprintf("https://%s%s", cfg.Domain, tokenPath),
		gracefulShutdownTimeout: time.Duration(cfg.GracefulShutdownTimeout) * time.Second,
	}

	dbInitialized, err := database.Initialized(server.boltdb)
	if err != nil {
		return nil, fmt.Errorf("unable to determine if the database has been initialized or not: %w", err)
	}

	server.dbInitialized = dbInitialized

	if err := server.setupRouter(); err != nil {
		return nil, fmt.Errorf("unable to set up the router: %w", err)
	}

	return &server, nil
}

func (s *Server) Serve() error {
	go func() {
		slog.Info(info.ApplicationName+" is now ready to serve web requests", "address", s.httpServer.Addr)

		if err := s.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server error", "error", err.Error())
			os.Exit(1)
		}
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	<-shutdownSignal

	slog.Info("Received the signal to shutdown Beacon.")

	ctx, cancel := context.WithTimeout(context.Background(), s.gracefulShutdownTimeout)
	defer cancel()

	if err := s.shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down Beacon: %w", err)
	}

	return nil
}

func (s *Server) shutdown(ctx context.Context) error {
	slog.Info("Shutting down the HTTP server.")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down the HTTP server: %w", err)
	}

	slog.Info("Closing the database.")

	if err := s.boltdb.Close(); err != nil {
		return fmt.Errorf("error closing the database: %w", err)
	}

	return nil
}

func (s *Server) setupRouter() error {
	mux := http.NewServeMux()

	staticRootFS, err := fs.Sub(staticFS, staticFSDir)
	if err != nil {
		return fmt.Errorf("unable to create the static root file system: %w", err)
	}

	fileServer := http.FileServer(http.FS(staticRootFS))

	mux.Handle("GET /setup", http.HandlerFunc(s.setup))
	mux.Handle("POST /setup", http.HandlerFunc(s.setup))
	mux.Handle("GET /static/", http.StripPrefix("/static", neuter(fileServer)))
	mux.Handle("GET /{$}", s.entrypoint(s.profileAuthorization(rootRedirect, s.profileRedirectToLogin)))
	mux.Handle("GET /.well-known/oauth-authorization-server", s.entrypoint(http.HandlerFunc(s.getMetadata)))
	mux.Handle("GET /profile/login", s.entrypoint(http.HandlerFunc(s.getLoginForm)))
	mux.Handle("POST /profile/login", s.entrypoint(http.HandlerFunc(s.authenticate)))
	mux.Handle("GET /profile/overview", s.entrypoint(s.profileAuthorization(s.getOverviewPage, s.profileRedirectToLogin)))
	mux.Handle("POST /profile/overview", s.entrypoint(s.profileAuthorization(s.updateProfileInformation, s.profileRedirectToLogin)))
	mux.Handle("POST /profile/logout", s.entrypoint(s.profileAuthorization(s.logout, s.profileRedirectToLogin)))
	mux.Handle("GET "+s.authPath, s.entrypoint(s.profileAuthorization(s.authorize, s.authorizeRedirectToLogin)))
	mux.Handle("POST "+s.authPath, s.entrypoint(s.exchangeAuthorization(s.profileExchange)))
	mux.Handle("POST "+s.authPath+"/accept", s.entrypoint(s.profileAuthorization(s.authorizeAccept, nil)))
	mux.Handle("POST "+s.authPath+"/reject", s.entrypoint(s.profileAuthorization(s.authorizeReject, nil)))
	mux.Handle("POST "+s.tokenPath, s.entrypoint(s.exchangeAuthorization(s.tokenExchange)))

	s.httpServer.Handler = mux

	return nil
}

func rootRedirect(writer http.ResponseWriter, request *http.Request, _ string) {
	http.Redirect(writer, request, "/profile/overview", http.StatusSeeOther)
}
