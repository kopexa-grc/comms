// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"

	"github.com/kopexa-grc/comms/driver"
)

type verifyEmailData struct {
	CommonData
	DisplayName string `json:"display_name"`
	Code        string `json:"code"`
}

func (c *Comms) SendVerifyEmail(ctx context.Context, r Recipient, code string) error {
	data := verifyEmailData{
		CommonData: CommonData{
			Subject:   "Verify your email address",
			Recipient: r,
		},
		DisplayName: r.Name(),
		Code:        code,
	}

	msg, err := c.newVerifyEmail(data)
	if err != nil {
		return err
	}

	return c.config.Driver.Send(ctx, msg)
}

func (c *Comms) newVerifyEmail(data verifyEmailData) (driver.Message, error) {
	text, html, err := Render("verify-email", data)
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
