// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"

	"github.com/kopexa-grc/comms/v2/driver"
)

type ReviewOverdueData struct {
	CommonData
	Entity     string `json:"entity"`
	EntityName string `json:"entity_name"`
	Space      string `json:"space"`
	URL        string `json:"url"`
}

func (c *Comms) SendReviewOverdueEmail(ctx context.Context, r Recipient, data ReviewOverdueData) error {
	data.CommonData = CommonData{
		Subject:   "Review overdue notification",
		Recipient: r,
	}

	msg, err := c.newReviewOverdueEmail(data)
	if err != nil {
		return err
	}

	return c.config.Driver.Send(ctx, msg)
}

func (c *Comms) newReviewOverdueEmail(data ReviewOverdueData) (driver.Message, error) {
	text, html, err := Render("review-overdue", data)
	if err != nil {
		return driver.Message{}, err
	}

	return driver.Message{
		From:    c.config.From,
		To:      []string{data.Recipient.String()},
		Subject: data.Subject,
		Text:    text,
		HTML:    html,
	}, nil
}
