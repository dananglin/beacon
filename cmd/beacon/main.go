// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"os"

	"codeflow.dananglin.me.uk/apollo/beacon/internal/executors"
)

func main() {
	if err := executors.Execute(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
