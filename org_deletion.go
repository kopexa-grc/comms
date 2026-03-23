// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"fmt"
	"html/template"

	"github.com/kopexa-grc/comms/v2/driver"
)

// OrganizationDeletionConfirmEmailData holds the data for the org deletion confirmation email.
type OrganizationDeletionConfirmEmailData struct {
	CommonData
	DisplayName      string
	OrganizationName string
	URL              template.URL
	ExpiresIn        string
}

// SendOrganizationDeletionConfirmEmail sends a confirmation email for organization deletion.
func (c *Comms) SendOrganizationDeletionConfirmEmail(ctx context.Context, r Recipient, data OrganizationDeletionConfirmEmailData) error {
	if err := r.Validate(); err != nil {
		return err
	}

	data.CommonData = CommonData{
		Subject:   fmt.Sprintf(Subject("org-deletion-confirm", r.Lang()), data.OrganizationName),
		Recipient: r,
	}
	data.DisplayName = r.Name()

	text, html, err := Render(r.Lang(), "org-deletion-confirm", data)
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
