// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"fmt"

	"github.com/kopexa-grc/comms/v2/driver"
)

type InviteEmailData struct {
	Actor        string
	ActorEmail   string
	Organization string
	Message      string
	URL          string
}

// SendInviteEmail sends an invitation email to a recipient to join an organization.
func (c *Comms) SendInviteEmail(ctx context.Context, recipient Recipient, data InviteEmailData) error {
	if err := recipient.Validate(); err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}

	text, html, err := Render("org-invite", map[string]string{
		"Actor":        data.Actor,
		"ActorEmail":   data.ActorEmail,
		"Organization": data.Organization,
		"Message":      data.Message,
		"Url":          data.URL,
		"DisplayName":  recipient.Name(),
	})
	if err != nil {
		return fmt.Errorf("failed to render invite email: %w", err)
	}

	message := driver.Message{
		From:    c.config.From,
		To:      []string{recipient.String()},
		Subject: fmt.Sprintf("%s has invited you to join %s on Kopexa", data.Actor, data.Organization),
		Text:    text,
		HTML:    html,
	}

	if err := c.config.Driver.Send(ctx, message); err != nil {
		return fmt.Errorf("failed to send invite email: %w", err)
	}

	return nil
}
