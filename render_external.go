// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"bytes"
	"fmt"
	"html"
	htmltemplate "html/template"
	"regexp"
	"strings"
	texttemplate "text/template"

	"github.com/Masterminds/sprig/v3"
)

var (
	reBlockClose = regexp.MustCompile(`</(?:p|div|h[1-6]|tr|li|blockquote)>`)
	reBr         = regexp.MustCompile(`<br\s*/?>`)
	reTag        = regexp.MustCompile(`<[^>]*>`)
	reSpaces     = regexp.MustCompile(`[^\S\n]+`)
	reBlankLines = regexp.MustCompile(`\n{3,}`)
)

// htmlToText converts an HTML string to plain text by stripping tags,
// converting block-level elements to newlines, and decoding HTML entities.
// Used as a fallback when no TextTemplate is provided.
func htmlToText(s string) string {
	if s == "" {
		return ""
	}

	// Convert <br> variants to newlines
	s = reBr.ReplaceAllString(s, "\n")

	// Convert closing block elements to newlines
	s = reBlockClose.ReplaceAllString(s, "\n")

	// Strip all remaining HTML tags
	s = reTag.ReplaceAllString(s, "")

	// Decode HTML entities
	s = html.UnescapeString(s)

	// Collapse horizontal whitespace (but preserve newlines)
	s = reSpaces.ReplaceAllString(s, " ")

	// Collapse 3+ newlines to 2
	s = reBlankLines.ReplaceAllString(s, "\n\n")

	return strings.TrimSpace(s)
}

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

// RenderTemplate renders an external template with branding and data into a ready-to-send email.
//
// The rendering pipeline:
//  1. Validate the template (at least one content template required)
//  2. Merge tmpl.Defaults with data (data takes precedence)
//  3. Resolve branding (fill zero values with Kopexa defaults)
//  4. Render SubjectTemplate and PreheaderTemplate with text/template
//  5. Render BodyTemplate with html/template (XSS protection)
//  6. Render TextTemplate with text/template (or auto-generate from body HTML)
//  7. Wrap rendered body in the branded email shell
//
// If branding is nil, Kopexa defaults are used.
// If both BodyTemplate and TextTemplate fail to render, an error is returned.
// If only one fails, the successful result is returned without error.
//
// Example:
//
//	tmpl := comms.ExternalTemplate{
//	    SubjectTemplate: "Welcome to {{.OrgName}}",
//	    BodyTemplate:    "<h1>Hello {{.Name}}</h1>",
//	    TextTemplate:    "Hello {{.Name}}",
//	}
//	rendered, err := comms.RenderTemplate(tmpl, nil, map[string]any{
//	    "Name":    "Max",
//	    "OrgName": "Acme Corp",
//	})
func RenderTemplate(tmpl ExternalTemplate, branding *Branding, data map[string]any) (*RenderedEmail, error) {
	if err := tmpl.Validate(); err != nil {
		return nil, err
	}

	resolved := resolveBranding(branding)
	merged := mergeData(tmpl.Defaults, data, resolved)

	// Render subject (text/template, no HTML escaping)
	subject, err := renderTextTmpl(tmpl.SubjectTemplate, merged)
	if err != nil {
		return nil, fmt.Errorf("failed to render subject: %w", err)
	}

	// Render preheader (text/template)
	preheader, err := renderTextTmpl(tmpl.PreheaderTemplate, merged)
	if err != nil {
		return nil, fmt.Errorf("failed to render preheader: %w", err)
	}

	// Render body (html/template for XSS protection)
	bodyHTML, bodyErr := renderHTMLTmpl(tmpl.BodyTemplate, merged)

	// Render plain text (text/template) or fall back to HTML-to-text conversion
	var textContent string

	var textErr error

	if tmpl.TextTemplate != "" {
		textContent, textErr = renderTextTmpl(tmpl.TextTemplate, merged)
	} else if bodyErr == nil && bodyHTML != "" {
		// Auto-generate plain text from rendered body HTML
		textContent = htmlToText(bodyHTML)
	}

	// Both failed → error
	if bodyErr != nil && textErr != nil {
		return nil, fmt.Errorf(
			"%w: body error: %v, text error: %v",
			ErrBothTemplatesFailed, bodyErr, textErr,
		)
	}

	// Wrap body in shell (only if body rendered successfully)
	var fullHTML string
	if bodyErr == nil && bodyHTML != "" {
		fullHTML, err = renderShell(bodyHTML, preheader, resolved)
		if err != nil {
			return nil, fmt.Errorf("failed to render shell: %w", err)
		}
	}

	return &RenderedEmail{
		Subject: subject,
		HTML:    fullHTML,
		Text:    textContent,
	}, nil
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
