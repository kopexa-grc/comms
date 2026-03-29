// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"errors"
	"fmt"

	"github.com/kopexa-grc/comms/v2/driver"
)

var (
	ErrRenderedRequired = errors.New("rendered email is required")
	ErrSubjectRequired  = errors.New("rendered email subject is required")
	ErrBodyRequired     = errors.New("rendered email must have HTML or Text body")
)

// SendRendered sends a pre-rendered email to the given recipient.
//
// Use this after RenderTemplate when you need to separate rendering from sending,
// e.g., for previews, caching, or bulk sends.
//
// Validates before sending:
//   - rendered must not be nil
//   - rendered.Subject must not be empty
//   - rendered.HTML or rendered.Text must be non-empty
//   - Recipient email must be valid
func (c *Comms) SendRendered(ctx context.Context, r Recipient, rendered *RenderedEmail) error {
	if rendered == nil {
		return fmt.Errorf("%w", ErrRenderedRequired)
	}

	if rendered.Subject == "" {
		return fmt.Errorf("%w", ErrSubjectRequired)
	}

	if rendered.HTML == "" && rendered.Text == "" {
		return fmt.Errorf("%w", ErrBodyRequired)
	}

	if err := r.Validate(); err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}

	msg := driver.Message{
		From:    c.config.From,
		To:      []string{r.String()},
		Subject: rendered.Subject,
		HTML:    rendered.HTML,
		Text:    rendered.Text,
	}

	if err := c.config.Driver.Send(ctx, msg); err != nil {
		return fmt.Errorf("failed to send rendered email: %w", err)
	}

	return nil
}

// SendFromTemplate is a convenience method that renders an external template and sends
// it in one call. Equivalent to calling RenderTemplate followed by SendRendered.
//
// Example:
//
//	tmpl := comms.ExternalTemplate{
//	    SubjectTemplate: "Welcome to {{.OrgName}}",
//	    BodyTemplate:    "<h1>Hello {{.Name}}</h1>",
//	    TextTemplate:    "Hello {{.Name}}",
//	}
//	err := c.SendFromTemplate(ctx, recipient, tmpl, nil, map[string]any{
//	    "Name":    "Max",
//	    "OrgName": "Acme Corp",
//	})
func (c *Comms) SendFromTemplate(ctx context.Context, r Recipient, tmpl ExternalTemplate, branding *Branding, data map[string]any) error {
	rendered, err := RenderTemplate(tmpl, branding, data)
	if err != nil {
		return err
	}

	return c.SendRendered(ctx, r, rendered)
}
