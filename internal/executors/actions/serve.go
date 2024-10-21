package actions

import (
	"flag"
	"fmt"

	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/server"
)

type Serve struct {
	*flag.FlagSet

	configPath string
}

func NewServe() *Serve {
	serve := Serve{
		FlagSet: flag.NewFlagSet("serve", flag.ExitOnError),
	}

	serve.StringVar(&serve.configPath, "config", "", "The path to the config file")

	return &serve
}

func (a *Serve) Execute(args []string) error {
	if err := a.Parse(args); err != nil {
		return fmt.Errorf("(%s) flag parsing error: %w", a.Name(), err)
	}

	srv, err := server.NewServer(a.configPath)
	if err != nil {
		return fmt.Errorf("unable to create the web server: %w", err)
	}

	if err := srv.Serve(); err != nil {
		return fmt.Errorf("web server error: %w", err)
	}

	return nil
}
