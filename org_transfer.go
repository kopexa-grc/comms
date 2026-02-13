// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"fmt"

	"github.com/kopexa-grc/comms/v2/driver"
)

// TransferSenderConfirmEmailData contains data for the sender confirmation email.
type TransferSenderConfirmEmailData struct {
	Organization  string
	ReceiverName  string
	ReceiverEmail string
	URL           string
}

// TransferReceiverInviteEmailData contains data for the receiver invitation email.
type TransferReceiverInviteEmailData struct {
	Organization string
	SenderName   string
	SenderEmail  string
	URL          string
}

// SendTransferSenderConfirmEmail sends an email to the sender to confirm the ownership transfer.
func (c *Comms) SendTransferSenderConfirmEmail(ctx context.Context, recipient Recipient, data TransferSenderConfirmEmailData) error {
	if err := recipient.Validate(); err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}

	text, html, err := Render("org-transfer-sender-confirm", map[string]string{
		"Organization":  data.Organization,
		"ReceiverName":  data.ReceiverName,
		"ReceiverEmail": data.ReceiverEmail,
		"Url":           data.URL,
		"DisplayName":   recipient.Name(),
	})
	if err != nil {
		return fmt.Errorf("failed to render transfer sender confirm email: %w", err)
	}

	message := driver.Message{
		From:    c.config.From,
		To:      []string{recipient.String()},
		Subject: fmt.Sprintf("Confirm ownership transfer for %s", data.Organization),
		Text:    text,
		HTML:    html,
	}

	if err := c.config.Driver.Send(ctx, message); err != nil {
		return fmt.Errorf("failed to send transfer sender confirm email: %w", err)
	}

	return nil
}

// SendTransferReceiverInviteEmail sends an email to the receiver inviting them to accept ownership.
func (c *Comms) SendTransferReceiverInviteEmail(ctx context.Context, recipient Recipient, data TransferReceiverInviteEmailData) error {
	if err := recipient.Validate(); err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}

	text, html, err := Render("org-transfer-receiver-invite", map[string]string{
		"Organization": data.Organization,
		"SenderName":   data.SenderName,
		"SenderEmail":  data.SenderEmail,
		"Url":          data.URL,
		"DisplayName":  recipient.Name(),
	})
	if err != nil {
		return fmt.Errorf("failed to render transfer receiver invite email: %w", err)
	}

	message := driver.Message{
		From:    c.config.From,
		To:      []string{recipient.String()},
		Subject: fmt.Sprintf("%s has invited you to become owner of %s", data.SenderName, data.Organization),
		Text:    text,
		HTML:    html,
	}

	if err := c.config.Driver.Send(ctx, message); err != nil {
		return fmt.Errorf("failed to send transfer receiver invite email: %w", err)
	}

	return nil
}
