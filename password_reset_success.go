// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"fmt"

	"github.com/kopexa-grc/comms/v2/driver"
)

// SendPasswordResetSuccessEmail sends a confirmation email to a recipient after their password has been successfully reset.
func (c *Comms) SendPasswordResetSuccessEmail(ctx context.Context, recipient Recipient) error {
	if err := recipient.Validate(); err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}

	text, html, err := Render("password-reset-success", map[string]string{
		"DisplayName": recipient.Name(),
	})
	if err != nil {
		return fmt.Errorf("failed to render password reset success email: %w", err)
	}

	message := driver.Message{
		From:    c.config.From,
		To:      []string{recipient.String()},
		Subject: "Your password has been successfully reset",
		Text:    text,
		HTML:    html,
	}

	if err := c.config.Driver.Send(ctx, message); err != nil {
		return fmt.Errorf("failed to send password reset success email: %w", err)
	}

	return nil
}
