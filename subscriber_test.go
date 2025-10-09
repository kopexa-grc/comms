// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"testing"

	"github.com/kopexa-grc/comms/v2/driver/mock"
	"github.com/stretchr/testify/require"
)

func TestComms_SendSubscriberEmail(t *testing.T) {
	recipient := Recipient{
		Email:     "inviter@example.com",
		FirstName: "Max",
		LastName:  "Mustermann",
	}
	token := "test-token"
	url := "https://app.kopexa.com/subscriber-verify?token=test-token"

	// Create Comms instance with mock driver
	comms := New(WithDriver(mock.NewDriver()))

	// Test sending
	err := comms.SendNewSubscriber(context.Background(), recipient, token)
	require.NoError(t, err)

	// Get the last sent message
	msg, ok := comms.config.Driver.(*mock.Driver).LastMessage()
	require.True(t, ok)
	require.NotNil(t, msg)

	// Verify message content
	require.Equal(t, "Thank you for subscribing", msg.Subject)
	require.Equal(t, []string{recipient.String()}, msg.To)

	// Verify text content
	require.NotEmpty(t, msg.Text)
	require.Contains(t, msg.Text, url)
}
