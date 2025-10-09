// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"fmt"

	"github.com/kopexa-grc/comms/v2/driver"
)

type InviteAcceptedData struct {
	InviteeName  string
	Organization string
}

// SendInviteAcceptedEmail sends a notification email to the inviter when their invitation has been accepted.
func (c *Comms) SendInviteAcceptedEmail(ctx context.Context, recipient Recipient, data InviteAcceptedData) error {
	if err := recipient.Validate(); err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}

	text, html, err := Render("org-invite-accepted", map[string]string{
		"DisplayName":  recipient.Name(),
		"InviteeName":  data.InviteeName,
		"Organization": data.Organization,
	})
	if err != nil {
		return fmt.Errorf("failed to render invite accepted email: %w", err)
	}

	message := driver.Message{
		From:    c.config.From,
		To:      []string{recipient.String()},
		Subject: fmt.Sprintf("%s has accepted your invitation to join %s", data.InviteeName, data.Organization),
		Text:    text,
		HTML:    html,
	}

	if err := c.config.Driver.Send(ctx, message); err != nil {
		return fmt.Errorf("failed to send invite accepted email: %w", err)
	}

	return nil
}
