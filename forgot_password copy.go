// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"fmt"

	"github.com/kopexa-grc/comms/v2/driver"
)

// SendForgotPasswordEmail sends a password reset email to a recipient.
// The email contains a one-time code that can be used to reset the password.
func (c *Comms) RecoveryCodesRegenerated(ctx context.Context, recipient Recipient, url string) error {
	if err := recipient.Validate(); err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}

	text, html, err := Render("recovery-codes-regenerated", map[string]string{
		"DisplayName": recipient.Name(),
		"URL":         url,
	})
	if err != nil {
		return fmt.Errorf("failed to render recovery codes regenerated email: %w", err)
	}

	message := driver.Message{
		From:    c.config.From,
		To:      []string{recipient.String()},
		Subject: "Reset your password",
		Text:    text,
		HTML:    html,
	}

	if err := c.config.Driver.Send(ctx, message); err != nil {
		return fmt.Errorf("failed to send recovery codes regenerated email: %w", err)
	}

	return nil
}
