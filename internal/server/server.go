// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/cache"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/config"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/database"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/info"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/ui"
	bolt "go.etcd.io/bbolt"
)

const (
	maxRequestSize int64 = 32 << 10 // 32KB

	activeTabSettings string = "settings"
	activeTabHome     string = "home"

	qKeyClientID            string = "client_id"
	qKeyCode                string = "code"
	qKeyCodeChallenge       string = "code_challenge"
	qKeyCodeChallengeMethod string = "code_challenge_method"
	qKeyError               string = "error"
	qKeyIssuer              string = "iss"
	qKeyLoginType           string = "login_type"
	qKeyMe                  string = "me"
	qKeyProfileID           string = "profile_id"
	qKeyRedirectURI         string = "redirect_uri"
	qKeyResponseType        string = "response_type"
	qKeyScope               string = "scope"
	qKeyState               string = "state"

	loginTypeProfile   string = "profile"
	loginTypeIndieauth string = "indieauth"

	pathAuth       string = "/indieauth/authorize"
	pathAuthAccept string = pathAuth + "/accept"
	pathAuthReject string = pathAuth + "/reject"
	pathToken      string = "/indieauth/token" // #nosec G101 -- This is not hardcoded credentials.
)

type (
	profileHandlerFunc  func(writer http.ResponseWriter, request *http.Request, profileID string)
	exchangeHandlerFunc func(writer http.ResponseWriter, data clientRequestData)

	Server struct {
		httpServer              *http.Server
		boltdb                  *bolt.DB
		cache                   *cache.Cache
		gracefulShutdownTimeout time.Duration
		htmlTemplate            *template.Template
		dbInitialized           bool
		domainName              string
		jwtSecret               string
		jwtCookieName           string
		authEndpoint            string
		issuer                  string
		tokenEndpoint           string
	}
)

func NewServer(configPath string) (*Server, error) {
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("error loading the configuration: %w", err)
	}

	boltdb, err := database.Open(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("error opening the database: %w", err)
	}

	tmpl, err := template.New("").ParseFS(ui.TemplateFS, ui.TemplatesDir+"/*")
	if err != nil {
		return nil, fmt.Errorf("error creating the HTML template: %w", err)
	}

	setupLogging(cfg.Log.Level)

	server := Server{
		httpServer: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port),
			ReadHeaderTimeout: 1 * time.Second,
		},
		boltdb:                  boltdb,
		cache:                   cache.NewCache(1 * time.Minute),
		gracefulShutdownTimeout: time.Duration(cfg.GracefulShutdownTimeout) * time.Second,
		htmlTemplate:            tmpl,
		domainName:              cfg.Domain,
		jwtSecret:               cfg.JWT.Secret,
		jwtCookieName:           cfg.JWT.CookieName,
		authEndpoint:            fmt.Sprintf("https://%s%s", cfg.Domain, pathAuth),
		issuer:                  fmt.Sprintf("https://%s/", cfg.Domain),
		tokenEndpoint:           fmt.Sprintf("https://%s%s", cfg.Domain, pathToken),
	}

	dbInitialized, err := database.Initialized(server.boltdb)
	if err != nil {
		return nil, fmt.Errorf(
			"error determining if the database has been initialized or not: %w",
			err,
		)
	}

	server.dbInitialized = dbInitialized

	server.setupRouter()

	return &server, nil
}

func (s *Server) Serve() error {
	go func() {
		slog.LogAttrs(
			context.Background(),
			slog.LevelInfo,
			info.ApplicationName+" is now ready to serve web requests",
			slog.String("address", s.httpServer.Addr),
		)

		if err := s.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.LogAttrs(
				context.Background(),
				slog.LevelError,
				"HTTP Server error",
				slog.Any("error", err),
			)

			os.Exit(1)
		}
	}()

	// Create the context for receiving the shutdown signal
	shutdownSignal, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	<-shutdownSignal.Done()
	stop()

	slog.LogAttrs(
		context.Background(),
		slog.LevelInfo,
		"Received the signal to shutdown Beacon.",
	)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		s.gracefulShutdownTimeout,
	)
	defer cancel()

	if err := s.shutdown(ctx); err != nil {
		slog.LogAttrs(
			context.Background(),
			slog.LevelError,
			"Error shutting down Beacon.",
			slog.Any("error", err),
		)

		return errors.New("server error")
	}

	return nil
}

func (s *Server) setupRouter() {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", neuter(http.FileServerFS(ui.StaticFS)))
	mux.Handle("GET /setup", s.entrypoint(http.HandlerFunc(s.setup)))
	mux.Handle("POST /setup", s.entrypoint(parseForm(s.setup)))
	mux.Handle("GET /{$}", s.entrypoint(s.profileAuthorization(redirectRoot, s.profileRedirectToLogin)))
	mux.Handle("GET /.well-known/oauth-authorization-server", s.entrypoint(http.HandlerFunc(s.getMetadata)))
	mux.Handle("GET /profile", s.entrypoint(http.HandlerFunc(s.redirectProfile)))
	mux.Handle("GET /profile/login", s.entrypoint(http.HandlerFunc(s.getLoginPage)))
	mux.Handle("POST /profile/login", s.entrypoint(parseForm(s.authenticate)))
	mux.Handle("GET /profile/overview", s.entrypoint(s.profileAuthorization(s.getOverviewPage, s.profileRedirectToLogin)))
	mux.Handle("POST /profile/logout", s.entrypoint(parseForm(s.profileAuthorization(s.logout, s.profileRedirectToLogin))))
	mux.Handle("GET /profile/settings", s.entrypoint(http.HandlerFunc(s.redirectProfileSettings)))
	mux.Handle("GET /profile/settings/info", s.entrypoint(s.profileAuthorization(s.getUpdateProfileInfoPage, s.profileRedirectToLogin)))
	mux.Handle("POST /profile/settings/info", s.entrypoint(parseForm(s.profileAuthorization(s.updateProfileInformation, s.profileRedirectToLogin))))
	mux.Handle("GET /profile/settings/password", s.entrypoint(s.profileAuthorization(s.getUpdatePasswordPage, s.profileRedirectToLogin)))
	mux.Handle("POST /profile/settings/password", s.entrypoint(parseForm(s.profileAuthorization(s.updateProfilePassword, s.profileRedirectToLogin))))
	mux.Handle("GET "+pathAuth, s.entrypoint(s.profileAuthorization(s.authorize, s.authorizeRedirectToLogin)))
	mux.Handle("POST "+pathAuth, s.entrypoint(parseForm(s.exchangeAuthorization(s.profileExchange))))
	mux.Handle("POST "+pathAuthAccept, s.entrypoint(parseForm(s.profileAuthorization(s.authorizeAccept, nil))))
	mux.Handle("POST "+pathAuthReject, s.entrypoint(parseForm(s.profileAuthorization(s.authorizeReject, nil))))
	mux.Handle("POST "+pathToken, s.entrypoint(parseForm(s.exchangeAuthorization(s.tokenExchange))))

	s.httpServer.Handler = mux
}

func (s *Server) shutdown(ctx context.Context) error {
	slog.LogAttrs(
		context.Background(),
		slog.LevelInfo,
		"Shutting down the HTTP server.",
	)

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down the HTTP server: %w", err)
	}

	slog.LogAttrs(
		context.Background(),
		slog.LevelInfo,
		"Closing the database.",
	)

	if err := s.boltdb.Close(); err != nil {
		return fmt.Errorf("error closing the database: %w", err)
	}

	return nil
}

func setupLogging(level string) {
	var logLevel slog.Level

	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := slog.HandlerOptions{
		Level: logLevel,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	slog.SetDefault(logger)
}

func logAccess(brw *responseWriter, requestID, clientAddr, userAgent, method, uri string) {
	slog.LogAttrs(
		context.Background(),
		slog.LevelInfo,
		http.StatusText(brw.statusCode)+": wrote "+strconv.Itoa(brw.wrote)+"B",
		slog.String("request_id", requestID),
		slog.String("client_addr", clientAddr),
		slog.String("user_agent", userAgent),
		slog.String("method", method),
		slog.String("uri", uri),
		slog.Int("status_code", brw.statusCode),
	)
}
