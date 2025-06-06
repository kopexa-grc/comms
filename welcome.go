// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"fmt"

	"github.com/kopexa-grc/comms/driver"
)

// SendWelcomeEmail sends a welcome email to a recipient after successful email verification.
// The email contains a personalized welcome message and confirms the successful verification.
func (c *Comms) SendWelcomeEmail(ctx context.Context, recipient Recipient) error {
	if err := recipient.Validate(); err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}

	text, html, err := Render("welcome", map[string]string{
		"DisplayName": recipient.Name(),
	})
	if err != nil {
		return fmt.Errorf("failed to render welcome email: %w", err)
	}

	message := driver.Message{
		From:    c.config.From,
		To:      []string{recipient.String()},
		Subject: "Welcome to Kopexa",
		Text:    text,
		HTML:    html,
	}

	if err := c.config.Driver.Send(ctx, message); err != nil {
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	return nil
}
