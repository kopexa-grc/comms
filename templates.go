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
	// DefaultLanguage is the fallback language when a template is not found
	// in the requested language.
	DefaultLanguage = "en"
)

var (
	//go:embed templates/en/* templates/de/*
	files embed.FS

	templates map[string]*template.Template
)

func init() {
	templates = make(map[string]*template.Template)

	// Read language directories
	langDirs, err := fs.ReadDir(files, defaultTemplatesDir)
	if err != nil {
		log.Panic().Err(err).Msg("could not read template directories")
	}

	for _, langDir := range langDirs {
		if !langDir.IsDir() {
			continue
		}

		lang := langDir.Name()
		langPath := filepath.Join(defaultTemplatesDir, lang)

		templateFiles, err := fs.ReadDir(files, langPath)
		if err != nil {
			log.Error().Err(err).Str("lang", lang).Msg("could not read language template directory")
			continue
		}

		for _, file := range templateFiles {
			if file.IsDir() {
				continue
			}

			// Key format: "en/verify-email.txt"
			key := filepath.Join(lang, file.Name())
			templates[key] = parseTemplate(key)
		}
	}
}

func parseTemplate(name string) *template.Template {
	// ParseFS names templates by their base filename only, not the full path.
	// So we need to use the base name for template.New() as well.
	baseName := filepath.Base(name)
	tmpl, err := template.New(baseName).Funcs(sprig.FuncMap()).ParseFS(files, filepath.Join(defaultTemplatesDir, name))
	if err != nil {
		log.Fatal().Err(err).Str("template", name).Msg("could not parse template")
	}

	return tmpl
}
