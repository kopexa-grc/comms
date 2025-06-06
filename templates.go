// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"embed"
	"io/fs"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/rs/zerolog/log"
)

const (
	defaultTemplatesDir = "templates"
)

var (
	//go:embed templates/*
	files embed.FS

	templates map[string]*template.Template
)

func init() {
	templates = make(map[string]*template.Template)

	templateFiles, err := fs.ReadDir(files, defaultTemplatesDir)
	if err != nil {
		log.Panic().Err(err).Msg("could not read template files")
	}

	for _, file := range templateFiles {
		if file.IsDir() {
			continue
		}

		templates[file.Name()] = parseTemplate(file.Name())
	}
}

func parseTemplate(name string) *template.Template {
	tmpl, err := template.New(name).Funcs(sprig.FuncMap()).ParseFS(files, filepath.Join(defaultTemplatesDir, name))
	if err != nil {
		log.Fatal().Err(err).Str("template", name).Msg("could not parse template")
	}

	return tmpl
}
