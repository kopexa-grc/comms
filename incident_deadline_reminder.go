// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"

	"github.com/kopexa-grc/comms/v2/driver"
)

// IncidentDeadlineReminderData contains data for the incident deadline reminder email
type IncidentDeadlineReminderData struct {
	CommonData
	IncidentTitle string `json:"incident_title"`
	IncidentID    string `json:"incident_id"`
	Framework     string `json:"framework"`     // e.g., "GDPR", "DORA", "NIS2"
	FrameworkName string `json:"framework_name"` // e.g., "General Data Protection Regulation"
	Deadline      string `json:"deadline"`      // formatted deadline time
	TimeLeft      string `json:"time_left"`     // e.g., "2 hours", "30 minutes"
	Space         string `json:"space"`
	URL           string `json:"url"`
}

// SendIncidentDeadlineReminderEmail sends a reminder email for an upcoming incident reporting deadline
func (c *Comms) SendIncidentDeadlineReminderEmail(ctx context.Context, r Recipient, data IncidentDeadlineReminderData) error {
	data.CommonData = CommonData{
		Subject:   "Incident reporting deadline approaching: " + data.IncidentTitle,
		Recipient: r,
	}

	msg, err := c.newIncidentDeadlineReminderEmail(data)
	if err != nil {
		return err
	}

	return c.config.Driver.Send(ctx, msg)
}

func (c *Comms) newIncidentDeadlineReminderEmail(data IncidentDeadlineReminderData) (driver.Message, error) {
	text, html, err := Render("incident-deadline-reminder", data)
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
