// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
)

var ErrBothTemplatesFailed = errors.New("both text and html templates failed")

// Render renders both text and HTML versions of a template for the given language.
// If the template is not found in the requested language, it falls back to English.
func Render(lang, name string, data any) (text, html string, err error) {
	if lang == "" {
		lang = DefaultLanguage
	}

	text, textErr := render(lang, name+".txt", data)
	html, htmlErr := render(lang, name+".html", data)

	if textErr != nil && htmlErr != nil {
		return "", "", fmt.Errorf("%w: text error: %v, html error: %v", ErrBothTemplatesFailed, textErr, htmlErr)
	}

	return text, html, nil
}

func render(lang, name string, data any) (string, error) {
	// Try requested language first
	key := filepath.Join(lang, name)
	t, ok := templates[key]

	// Fallback to default language if not found
	if !ok && lang != DefaultLanguage {
		key = filepath.Join(DefaultLanguage, name)
		t, ok = templates[key]
	}

	if !ok {
		return "", fmt.Errorf("%w: %q not found", ErrTemplateNotFound, name)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
