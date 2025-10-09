// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"testing"

	"github.com/kopexa-grc/comms/v2/driver/mock"
	"github.com/stretchr/testify/require"
)

func TestComms_SendForgotPasswordEmail(t *testing.T) {
	// Test data
	recipient := Recipient{
		Email:     "test@example.com",
		FirstName: "Max",
		LastName:  "Mustermann",
	}
	url := "http://localhost:3000/auth/password-reset?token=FGzm0xBF2HzdaGfql0HHqzm2PRDzPOQHSgW2fFQ3n6M"

	// Create Comms instance with mock driver
	comms := New(WithDriver(mock.NewDriver()))

	// Test sending
	err := comms.SendForgotPasswordEmail(context.Background(), recipient, url)
	require.NoError(t, err)

	// Get the last sent message
	msg, ok := comms.config.Driver.(*mock.Driver).LastMessage()
	require.True(t, ok)
	require.NotNil(t, msg)

	// Verify message content
	require.Equal(t, "Reset your password", msg.Subject)
	require.Equal(t, []string{recipient.String()}, msg.To)

	// Verify text content
	require.NotEmpty(t, msg.Text)
	require.Contains(t, msg.Text, recipient.Name())
	require.Contains(t, msg.Text, url)

	// Verify HTML content
	require.NotEmpty(t, msg.HTML)
	require.Contains(t, msg.HTML, recipient.Name())
	require.Contains(t, msg.HTML, url)
	require.Contains(t, msg.HTML, "<!DOCTYPE html PUBLIC")
}

func TestComms_SendForgotPasswordEmail_InvalidRecipient(t *testing.T) {
	// Test with invalid recipient
	recipient := Recipient{
		Email: "", // Invalid email
	}
	code := "123456"

	// Create Comms instance with mock driver
	comms := New(WithDriver(mock.NewDriver()))

	// Test sending
	err := comms.SendForgotPasswordEmail(context.Background(), recipient, code)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid recipient")
}
