// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"

	"github.com/kopexa-grc/comms/v2/driver"
)

// DsrReceiptConfirmationData contains data for the DSR receipt confirmation email
type DsrReceiptConfirmationData struct {
	CommonData
	// OrganizationName is the name of the organization processing the request
	OrganizationName string `json:"organization_name"`
	// DisplayID is the reference number for the request (e.g., "DSR-00042")
	DisplayID string `json:"display_id"`
	// AffectedPersonName is the full name of the data subject
	AffectedPersonName string `json:"affected_person_name"`
	// RequestTypes is a human-readable list of request types (e.g., "Data Access, Data Deletion")
	RequestTypes string `json:"request_types"`
	// ReceivedAt is the formatted date when the request was received
	ReceivedAt string `json:"received_at"`
	// ContactEmail is the email address for inquiries about the request
	ContactEmail string `json:"contact_email"`
}

// SendDsrReceiptConfirmationEmail sends a confirmation email to the data subject
// acknowledging receipt of their data subject request
func (c *Comms) SendDsrReceiptConfirmationEmail(ctx context.Context, r Recipient, data DsrReceiptConfirmationData) error {
	data.CommonData = CommonData{
		Subject:   "Confirmation of Your Data Subject Request - " + data.DisplayID,
		Recipient: r,
	}

	msg, err := c.newDsrReceiptConfirmationEmail(data)
	if err != nil {
		return err
	}

	return c.config.Driver.Send(ctx, msg)
}

func (c *Comms) newDsrReceiptConfirmationEmail(data DsrReceiptConfirmationData) (driver.Message, error) {
	text, html, err := Render("dsr-receipt-confirmation", data)
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
