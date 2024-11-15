// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/cache"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/config"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/info"
	bolt "go.etcd.io/bbolt"
)

type Server struct {
	httpServer          *http.Server
	boltdb              *bolt.DB
	cache               *cache.Cache
	domainName          string
	jwtSecret           string
	jwtCookieName       string
	indieauthCookieName string
	authPath            string
	authEndpoint        string
	issuer              string
	tokenEndpoint       string
	tokenPath           string
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

	authPath := "/indieauth/authorize"
	tokenPath := "/indieauth/token" // #nosec G101 -- This is not hardcoded credentials.

	server := Server{
		httpServer: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port),
			ReadHeaderTimeout: 1 * time.Second,
		},
		boltdb:              boltdb,
		cache:               cache.NewCache(1 * time.Minute),
		domainName:          cfg.Domain,
		jwtSecret:           cfg.JWT.Secret,
		jwtCookieName:       "beacon_is_great",
		authPath:            authPath,
		authEndpoint:        fmt.Sprintf("https://%s%s", cfg.Domain, authPath),
		indieauthCookieName: "indieauth_is_great",
		issuer:              fmt.Sprintf("https://%s/", cfg.Domain),
		tokenPath:           tokenPath,
		tokenEndpoint:       fmt.Sprintf("https://%s%s", cfg.Domain, tokenPath),
	}

	if err := server.setupRouter(); err != nil {
		return nil, fmt.Errorf("unable to set up the router: %w", err)
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

func (s *Server) setupRouter() error {
	mux := http.NewServeMux()

	staticRootFS, err := fs.Sub(staticFS, staticFSDir)
	if err != nil {
		return fmt.Errorf("unable to create the static root file system: %w", err)
	}

	fileServer := http.FileServer(http.FS(staticRootFS))

	mux.Handle("GET /{$}", http.HandlerFunc(s.rootRedirect))
	mux.Handle("GET /static/", http.StripPrefix("/static", neuter(fileServer)))
	mux.Handle("GET /.well-known/oauth-authorization-server", setRequestID(http.HandlerFunc(s.getMetadata)))
	mux.Handle("GET /profile/login", setRequestID(http.HandlerFunc(s.getLoginForm)))
	mux.Handle("POST /profile/login", setRequestID(http.HandlerFunc(s.authenticate)))
	mux.Handle("GET /profile/overview", setRequestID(s.protected(s.getOverviewPage, s.profileRedirectToLogin)))
	mux.Handle("POST /profile/overview", setRequestID(s.protected(s.updateProfileInformation, s.profileRedirectToLogin)))
	mux.Handle("GET /profile/setup", setRequestID(http.HandlerFunc(s.setup)))
	mux.Handle("POST /profile/setup", setRequestID(http.HandlerFunc(s.setup)))
	mux.Handle("GET "+s.authPath, setRequestID(s.protected(s.authorize, s.authorizeRedirectToLogin)))
	mux.Handle("POST "+s.authPath+"/accept", setRequestID(s.protected(s.authorizeAccept, nil)))
	mux.Handle("POST "+s.authPath+"/reject", setRequestID(s.protected(s.authorizeReject, nil)))

	s.httpServer.Handler = mux

	return nil
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

func neuter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasSuffix(request.URL.Path, "/") {
			sendClientError(
				writer,
				http.StatusNotFound,
				fmt.Errorf("the path must not end with a %q", "/"),
			)
		}

		next.ServeHTTP(writer, request)
	})
}
