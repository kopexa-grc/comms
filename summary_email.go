// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"

	"github.com/kopexa-grc/comms/v2/driver"
)

// SummaryEntityItem represents a single actionable item.
type SummaryEntityItem struct {
	EntityName string `json:"entity_name"`
	EntityURL  string `json:"entity_url"`
}

// SummaryEntityGroup groups items by entity type (e.g. "Risiken").
// Label is already singular/plural resolved.
type SummaryEntityGroup struct {
	Label string              `json:"label"`
	Items []SummaryEntityItem `json:"items"`
}

// SummarySpaceGroup groups entity types by space within a category.
type SummarySpaceGroup struct {
	SpaceName    string               `json:"space_name"`
	EntityGroups []SummaryEntityGroup `json:"entity_groups"`
}

// SummaryCategorySection is the top-level grouping (e.g. "Überfällige Reviews").
type SummaryCategorySection struct {
	Label  string              `json:"label"` // e.g. "Overdue Reviews"
	Spaces []SummarySpaceGroup `json:"spaces"`
}

// SummaryEmailData holds the full data for the daily summary email.
type SummaryEmailData struct {
	CommonData
	Sections   []SummaryCategorySection `json:"sections"`
	TotalCount int                      `json:"total_count"`
}

func (c *Comms) SendSummaryEmail(ctx context.Context, r Recipient, data SummaryEmailData) error {
	data.CommonData = CommonData{
		Subject:   Subject("summary", r.Lang()),
		Recipient: r,
	}

	msg, err := c.newSummaryEmail(data)
	if err != nil {
		return err
	}

	return c.config.Driver.Send(ctx, msg)
}

func (c *Comms) newSummaryEmail(data SummaryEmailData) (driver.Message, error) {
	text, html, err := Render(data.Recipient.Lang(), "summary", data)
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
