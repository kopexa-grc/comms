// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"testing"

	"github.com/kopexa-grc/comms/v2/driver/mock"
	"github.com/stretchr/testify/require"
)

func TestComms_SendOrgCreatedEmail(t *testing.T) {
	// Test data
	recipient := Recipient{
		Email:     "test@example.com",
		FirstName: "Max",
		LastName:  "Mustermann",
	}
	orgName := "Test Organization"

	// Create Comms instance with mock driver
	comms := New(WithDriver(mock.NewDriver()))

	// Test sending
	err := comms.SendOrgCreatedEmail(context.Background(), recipient, orgName)
	require.NoError(t, err)

	// Get the last sent message
	msg, ok := comms.config.Driver.(*mock.Driver).LastMessage()
	require.True(t, ok)
	require.NotNil(t, msg)

	// Verify message content
	require.Equal(t, "Your organization has been created", msg.Subject)
	require.Equal(t, []string{recipient.String()}, msg.To)

	// Verify text content
	require.NotEmpty(t, msg.Text)
	require.Contains(t, msg.Text, recipient.Name())
	require.Contains(t, msg.Text, orgName)

	// Verify HTML content
	require.NotEmpty(t, msg.HTML)
	require.Contains(t, msg.HTML, recipient.Name())
	require.Contains(t, msg.HTML, orgName)
	require.Contains(t, msg.HTML, "<!DOCTYPE html PUBLIC")
}

func TestComms_SendOrgCreatedEmail_InvalidRecipient(t *testing.T) {
	// Test with invalid recipient
	recipient := Recipient{
		Email: "", // Invalid email
	}
	orgName := "Test Organization"

	// Create Comms instance with mock driver
	comms := New(WithDriver(mock.NewDriver()))

	// Test sending
	err := comms.SendOrgCreatedEmail(context.Background(), recipient, orgName)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid recipient")
}
