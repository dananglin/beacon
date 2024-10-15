package executors

import (
	"flag"
	"fmt"

	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/config"
	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/server"
)

type serveExecutor struct {
	*flag.FlagSet

	configFile string
}

func executeServeCommand(args []string) error {
	executorName := "serve"

	executor := serveExecutor{
		FlagSet: flag.NewFlagSet(executorName, flag.ExitOnError),
	}

	executor.StringVar(&executor.configFile, "config", "", "The path to the config file")

	if err := executor.Parse(args); err != nil {
		return fmt.Errorf("(%s) flag parsing error: %w", executorName, err)
	}

	if err := executor.execute(); err != nil {
		return fmt.Errorf("(%s) execution error: %w", executorName, err)
	}

	return nil
}

func (e *serveExecutor) execute() error {
	cfg, err := config.NewConfig(e.configFile)
	if err != nil {
		return fmt.Errorf("unable to load the config: %w", err)
	}

	srv := server.NewServer(cfg)

	return srv.Serve()
}
