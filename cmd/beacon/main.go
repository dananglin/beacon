// SPDX-FileCopyrightText: 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"log/slog"
	"os"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/executors"
)

func main() {
	// Set up logging
	loggingLevel := new(slog.LevelVar)

	slogOpts := slog.HandlerOptions{Level: loggingLevel}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slogOpts))
	slog.SetDefault(logger)
	loggingLevel.Set(slog.LevelInfo)

	if err := executors.Execute(os.Args[1:]); err != nil {
		slog.Error(err.Error())
	}
}
