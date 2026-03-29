// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import "errors"

var ErrTemplateContentRequired = errors.New("at least one of BodyTemplate or TextTemplate is required")

// ExternalTemplate holds template strings loaded from an external source (e.g., database).
// Templates use Go's html/template syntax with Sprig v3 functions available.
//
// The BodyTemplate is rendered inside the email shell — do NOT include
// DOCTYPE, html, or body tags. The shell provides the email chrome
// (header, footer, branding) automatically.
//
// Template variables are passed as map[string]any. Branding values are
// accessible via {{.Branding.ButtonColor}}, {{.Branding.LinkColor}}, etc.
type ExternalTemplate struct {
	// SubjectTemplate is a Go template string for the email subject line.
	// Rendered with text/template (no HTML escaping).
	// Example: "{{.ActorName}} invited you to {{.OrgName}}"
	SubjectTemplate string

	// PreheaderTemplate is a Go template string for the email preview text.
	// This appears in email clients before opening the email.
	// Rendered with text/template (no HTML escaping).
	PreheaderTemplate string

	// BodyTemplate is a Go html/template string for the email body content.
	// Rendered inside the email shell — do NOT include DOCTYPE/html/body tags.
	BodyTemplate string

	// TextTemplate is a Go text/template string for the plain text fallback.
	// If empty, no plain text version is included in the rendered email.
	TextTemplate string

	// Defaults are base-layer variables merged with call-site data.
	// Call-site data takes precedence over defaults.
	Defaults map[string]any
}

// Validate checks that the ExternalTemplate has at least one content template.
// Returns ErrTemplateContentRequired if both BodyTemplate and TextTemplate are empty.
func (t *ExternalTemplate) Validate() error {
	if t.BodyTemplate == "" && t.TextTemplate == "" {
		return ErrTemplateContentRequired
	}

	return nil
}

// RenderedEmail is the result of rendering an external template with branding.
// It contains the fully rendered subject, HTML body, and plain text body
// ready for sending via SendRendered.
type RenderedEmail struct {
	// Subject is the rendered email subject line.
	Subject string

	// HTML is the complete HTML email document including the branded shell.
	HTML string

	// Text is the rendered plain text version of the email.
	Text string
}
