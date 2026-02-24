// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"

	"github.com/kopexa-grc/comms/v2/driver"
)

// IncidentDeadlineOverdueData contains data for the incident deadline overdue email
type IncidentDeadlineOverdueData struct {
	CommonData
	IncidentTitle string `json:"incident_title"`
	IncidentID    string `json:"incident_id"`
	Framework     string `json:"framework"`     // e.g., "GDPR", "DORA", "NIS2"
	FrameworkName string `json:"framework_name"` // e.g., "General Data Protection Regulation"
	Deadline      string `json:"deadline"`      // formatted deadline time
	TimeOverdue   string `json:"time_overdue"`  // e.g., "2 hours", "1 day"
	Space         string `json:"space"`
	URL           string `json:"url"`
}

// SendIncidentDeadlineOverdueEmail sends a notification email when an incident reporting deadline has passed
func (c *Comms) SendIncidentDeadlineOverdueEmail(ctx context.Context, r Recipient, data IncidentDeadlineOverdueData) error {
	data.CommonData = CommonData{
		Subject:   "URGENT: Incident reporting deadline overdue: " + data.IncidentTitle,
		Recipient: r,
	}

	msg, err := c.newIncidentDeadlineOverdueEmail(data)
	if err != nil {
		return err
	}

	return c.config.Driver.Send(ctx, msg)
}

func (c *Comms) newIncidentDeadlineOverdueEmail(data IncidentDeadlineOverdueData) (driver.Message, error) {
	text, html, err := Render("incident-deadline-overdue", data)
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
