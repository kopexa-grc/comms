// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"testing"

	"github.com/kopexa-grc/comms/v2/driver/mock"
	"github.com/stretchr/testify/require"
)

func TestComms_SendVendorAssessmentEmail(t *testing.T) {
	recipient := Recipient{
		Email:     "vendor@example.com",
		FirstName: "Max",
		LastName:  "Mustermann",
	}
	data := VendorAssessmentData{
		ActorName:        "Alice Example",
		OrganizationName: "Acme GmbH",
		AssessmentURL:    "https://app.kopexa.com/assessment/123",
	}

	comms := New(WithDriver(mock.NewDriver()))

	err := comms.SendVendorAssessmentEmail(context.Background(), recipient, data)
	require.NoError(t, err)

	msg, ok := comms.config.Driver.(*mock.Driver).LastMessage()
	require.True(t, ok)
	require.NotNil(t, msg)

	require.Contains(t, msg.Subject, data.OrganizationName)
	require.Equal(t, []string{recipient.String()}, msg.To)

	require.NotEmpty(t, msg.Text)
	require.Contains(t, msg.Text, data.ActorName)
	require.Contains(t, msg.Text, data.OrganizationName)
	require.Contains(t, msg.Text, data.AssessmentURL)
	require.Contains(t, msg.Text, "Vendor Assessment Request")

	require.NotEmpty(t, msg.HTML)
	require.Contains(t, msg.HTML, data.ActorName)
	require.Contains(t, msg.HTML, data.OrganizationName)
	require.Contains(t, msg.HTML, data.AssessmentURL)
	require.Contains(t, msg.HTML, "<!DOCTYPE html PUBLIC")
}

func TestComms_SendVendorAssessmentEmail_InvalidRecipient(t *testing.T) {
	recipient := Recipient{Email: ""}
	data := VendorAssessmentData{
		ActorName:        "Alice Example",
		OrganizationName: "Acme GmbH",
		AssessmentURL:    "https://app.kopexa.com/assessment/123",
	}
	comms := New(WithDriver(mock.NewDriver()))
	err := comms.SendVendorAssessmentEmail(context.Background(), recipient, data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid recipient")
}
