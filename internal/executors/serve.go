package executors

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/info"
	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/router"
)

type serveExecutor struct {
	*flag.FlagSet

	address string
}

func executeServeCommand(args []string) error {
	executorName := "serve"

	executor := serveExecutor{
		FlagSet: flag.NewFlagSet(executorName, flag.ExitOnError),
	}

	executor.StringVar(&executor.address, "address", "0.0.0.0:8080", "The address that the server will listen on")

	if err := executor.Parse(args); err != nil {
		return fmt.Errorf("(%s) flag parsing error: %w", executorName, err)
	}

	if err := executor.execute(); err != nil {
		return fmt.Errorf("(%s) execution error: %w", executorName, err)
	}

	return nil
}

func (e *serveExecutor) execute() error {
	server := http.Server{
		Addr:              e.address,
		Handler:           router.NewServeMux(),
		ReadHeaderTimeout: 1 * time.Second,
	}

	slog.Info(info.ApplicationName+" is listening for web requests", "address", e.address)

	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("error running the server: %w", err)
	}

	return nil
}
