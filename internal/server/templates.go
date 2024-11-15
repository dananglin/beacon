// SPDX-FileCopyrightText: 2024 2024 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package server

import "embed"

const (
	templatesFSDir = "ui/templates"
	staticFSDir    = "ui/static"
)

//go:embed ui/templates/*
var templatesFS embed.FS

//go:embed ui/static/*
var staticFS embed.FS
