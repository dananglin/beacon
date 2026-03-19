// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package executors

import (
	"errors"
	"os"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/executors/actions"
	"codeflow.dananglin.me.uk/apollo/beacon/internal/utilities"
)

var (
	errExecution  = errors.New("execution error")
	errArgParsing = errors.New("args parsing error")
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
		_, _ = os.Stderr.WriteString("ERROR: args parsing error: " + err.Error() + "\n")

		return errArgParsing
	}

	action, ok := actionMap[actionArgs.Name]
	if !ok {
		err := UnrecognisedActionError{action: actionArgs.Name}

		_, _ = os.Stderr.WriteString("ERROR: " + err.Error() + "\n")

		return err
	}

	if err := action.Execute(actionArgs.Args); err != nil {
		_, _ = os.Stderr.WriteString("ERROR: execution error: " + err.Error() + "\n")

		return errExecution
	}

	return nil
}
