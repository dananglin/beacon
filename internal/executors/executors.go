package executors

import (
	"fmt"

	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/executors/actions"
	"codeflow.dananglin.me.uk/apollo/indieauth-server/internal/utilities"
)

type UnrecognisedActionError struct {
	action string
}

func (e UnrecognisedActionError) Error() string {
	return "unrecognised action: " + e.action
}

func Execute(args []string) error {
	actionMap := map[string]actions.Executor{
		"serve":   actions.NewServe(),
		"version": actions.NewVersion(),
	}

	actionArgs, err := utilities.ParseArgs(args)
	if err != nil {
		return fmt.Errorf("args parsing error: %w", err)
	}

	action, ok := actionMap[actionArgs.Name]
	if !ok {
		return UnrecognisedActionError{action: actionArgs.Name}
	}

	if err := action.Execute(actionArgs.Args); err != nil {
		return fmt.Errorf("execution error: %w", err)
	}

	return nil
}
