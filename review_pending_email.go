// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"

	"github.com/kopexa-grc/comms/v2/driver"
)

type ReviewPendingData struct {
	CommonData
	Entity       string `json:"entity"`
	EntityName   string `json:"entity_name"`
	Space        string `json:"space"`
	URL          string `json:"url"`
	NextReviewAt string `json:"next_review_at"`
}

func (c *Comms) SendReviewPendingEmail(ctx context.Context, r Recipient, data ReviewPendingData) error {
	data.CommonData = CommonData{
		Subject:   Subject("review-pending", r.Lang()),
		Recipient: r,
	}

	msg, err := c.newReviewPendingEmail(data)
	if err != nil {
		return err
	}

	return c.config.Driver.Send(ctx, msg)
}

func (c *Comms) newReviewPendingEmail(data ReviewPendingData) (driver.Message, error) {
	text, html, err := Render(data.Recipient.Lang(), "review-pending", data)
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
