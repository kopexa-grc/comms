// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"

	"github.com/kopexa-grc/comms/driver"
)

type subscriberEmailData struct {
	CommonData
	URL string `json:"url"`
}

const subscriptionBaseURL = "https://app.kopexa.com/subscriber-verify"

func (c *Comms) SendNewSubscriber(ctx context.Context, r Recipient, token string) error {
	var err error

	data := subscriberEmailData{
		CommonData: CommonData{
			Subject:   "Thank you for subscribing",
			Recipient: r,
		},
	}

	data.URL, err = addTokenToURL(subscriptionBaseURL, token)
	if err != nil {
		return err
	}

	msg, err := c.newSubscriberEmail(data)
	if err != nil {
		return err
	}

	return c.config.Driver.Send(ctx, msg)
}

func (c *Comms) newSubscriberEmail(data subscriberEmailData) (driver.Message, error) {
	text, html, err := Render("subscribe", data)
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
