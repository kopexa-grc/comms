// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"fmt"

	"github.com/kopexa-grc/comms/v2/driver"
)

// SendSurveyOTP sends a one-time password email to the survey recipient.
func (c *Comms) SendSurveyOTP(ctx context.Context, recipientEmail, recipientName, otpCode string) error {
	recipient := Recipient{
		Email:     recipientEmail,
		FirstName: recipientName,
	}

	if err := recipient.Validate(); err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}

	text, html, err := Render(recipient.Lang(), "vendor-survey-otp", map[string]string{
		"DisplayName": recipient.Name(),
		"OTPCode":     otpCode,
		"ExpiresIn":   "5 minutes",
	})
	if err != nil {
		return fmt.Errorf("failed to render survey OTP email: %w", err)
	}

	message := driver.Message{
		From:    c.config.From,
		To:      []string{recipient.String()},
		Subject: Subject("vendor-survey-otp", recipient.Lang()),
		Text:    text,
		HTML:    html,
	}

	if err := c.config.Driver.Send(ctx, message); err != nil {
		return fmt.Errorf("failed to send survey OTP email: %w", err)
	}

	return nil
}
