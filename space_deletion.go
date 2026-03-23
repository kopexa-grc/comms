// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"fmt"
	"html/template"

	"github.com/kopexa-grc/comms/v2/driver"
)

// SpaceDeletionConfirmEmailData holds the data for the space deletion confirmation email.
type SpaceDeletionConfirmEmailData struct {
	CommonData
	DisplayName      string
	SpaceName        string
	OrganizationName string
	URL              template.URL
	ExpiresIn        string
}

// SendSpaceDeletionConfirmEmail sends a confirmation email for space deletion.
func (c *Comms) SendSpaceDeletionConfirmEmail(ctx context.Context, r Recipient, data SpaceDeletionConfirmEmailData) error {
	if err := r.Validate(); err != nil {
		return err
	}

	data.CommonData = CommonData{
		Subject:   fmt.Sprintf(Subject("space-deletion-confirm", r.Lang()), data.SpaceName),
		Recipient: r,
	}
	data.DisplayName = r.Name()

	text, html, err := Render(r.Lang(), "space-deletion-confirm", data)
	if err != nil {
		return err
	}

	msg := driver.Message{
		From:    c.config.From,
		To:      []string{r.String()},
		Subject: data.Subject,
		Text:    text,
		HTML:    html,
	}

	return c.config.Driver.Send(ctx, msg)
}
