// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"fmt"

	"github.com/kopexa-grc/comms/v2/driver"
)

type VendorAssessmentData struct {
	ActorName        string
	OrganizationName string
	AssessmentURL    string
}

// SendVendorAssessmentEmail sends a request email to a vendor to complete an assessment.
func (c *Comms) SendVendorAssessmentEmail(ctx context.Context, recipient Recipient, data VendorAssessmentData) error {
	if err := recipient.Validate(); err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}

	text, html, err := Render(recipient.Lang(), "vendor-assessment-request", map[string]string{
		"DisplayName":      recipient.Name(),
		"ActorName":        data.ActorName,
		"OrganizationName": data.OrganizationName,
		"AssessmentUrl":    data.AssessmentURL,
	})
	if err != nil {
		return fmt.Errorf("failed to render vendor assessment email: %w", err)
	}

	message := driver.Message{
		From:    c.config.From,
		To:      []string{recipient.String()},
		Subject: fmt.Sprintf(Subject("vendor-assessment-request", recipient.Lang()), data.OrganizationName),
		Text:    text,
		HTML:    html,
	}

	if err := c.config.Driver.Send(ctx, message); err != nil {
		return fmt.Errorf("failed to send vendor assessment email: %w", err)
	}

	return nil
}
