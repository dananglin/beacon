// SPDX-FileCopyrightText: 2026 Dan Anglin <d.n.i.anglin@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package ui

import "embed"

const (
	TemplatesDir = "templates"
	StaticDir    = "static"
)

//go:embed "static"
var StaticFS embed.FS

//go:embed "all:templates"
var TemplateFS embed.FS
