// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	texttemplate "text/template"

	"github.com/Masterminds/sprig/v3"
)

// mergeData merges template defaults with call-site data.
// Call-site data takes precedence over defaults.
// The Branding key is always set from the resolved branding and cannot be overridden.
func mergeData(defaults, data map[string]any, branding Branding) map[string]any {
	merged := make(map[string]any)

	for k, v := range defaults {
		merged[k] = v
	}

	for k, v := range data {
		merged[k] = v
	}

	// Branding is always injected and cannot be overridden
	merged["Branding"] = branding

	return merged
}

// renderTextTmpl renders a Go text/template string with data and Sprig functions.
// Returns empty string without error if tmplStr is empty.
func renderTextTmpl(tmplStr string, data any) (string, error) {
	if tmplStr == "" {
		return "", nil
	}

	t, err := texttemplate.New("").Option("missingkey=error").Funcs(sprig.TxtFuncMap()).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse text template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute text template: %w", err)
	}

	return buf.String(), nil
}

// renderHTMLTmpl renders a Go html/template string with data and Sprig functions.
// HTML-escapes values by default for XSS protection.
// Returns empty string without error if tmplStr is empty.
func renderHTMLTmpl(tmplStr string, data any) (string, error) {
	if tmplStr == "" {
		return "", nil
	}

	t, err := htmltemplate.New("").Option("missingkey=error").Funcs(sprig.HtmlFuncMap()).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse html template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute html template: %w", err)
	}

	return buf.String(), nil
}

// shellData holds the data passed to the email shell template.
type shellData struct {
	Body      htmltemplate.HTML
	Preheader string
	Branding  Branding
}

// renderShell wraps rendered body HTML in the branded email shell.
// The body parameter is treated as pre-rendered trusted HTML.
func renderShell(body, preheader string, branding Branding) (string, error) {
	t, err := htmltemplate.New("shell").Parse(emailShellTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse shell template: %w", err)
	}

	data := shellData{
		Body:      htmltemplate.HTML(body), // #nosec G203 — body is pre-rendered by renderHTMLTmpl
		Preheader: preheader,
		Branding:  branding,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute shell template: %w", err)
	}

	return buf.String(), nil
}
