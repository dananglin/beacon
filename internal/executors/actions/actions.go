// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package actions

type UnrecognisedResouceError struct {
	resource string
}

func (e UnrecognisedResouceError) Error() string {
	return "unrecognised resource: " + e.resource
}

type Executor interface {
	Execute(args []string) error
}
