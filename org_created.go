// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"fmt"

	"github.com/kopexa-grc/comms/v2/driver"
)

// SendOrgCreatedEmail sends a notification email to the organization creator
// after successful organization creation.
func (c *Comms) SendOrgCreatedEmail(ctx context.Context, recipient Recipient, orgName string) error {
	if err := recipient.Validate(); err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}

	text, html, err := Render("org-created", map[string]string{
		"DisplayName": recipient.Name(),
		"OrgName":     orgName,
		"ORGNAME":     orgName, // text compatibility
	})
	if err != nil {
		return fmt.Errorf("failed to render org created email: %w", err)
	}

	message := driver.Message{
		From:    c.config.From,
		To:      []string{recipient.String()},
		Subject: "Your organization has been created",
		Text:    text,
		HTML:    html,
	}

	if err := c.config.Driver.Send(ctx, message); err != nil {
		return fmt.Errorf("failed to send org created email: %w", err)
	}

	return nil
}
