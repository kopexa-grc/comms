// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"testing"

	"github.com/kopexa-grc/comms/v2/driver/mock"
	"github.com/stretchr/testify/require"
)

func TestComms_SendInviteEmail(t *testing.T) {
	recipient := Recipient{
		Email:     "invitee@example.com",
		FirstName: "Anna",
		LastName:  "Musterfrau",
	}
	data := InviteEmailData{
		Actor:        "Max Mustermann",
		ActorEmail:   "max@example.com",
		Organization: "TestOrg",
		Message:      "Let's build something great together!",
		URL:          "https://kopexa.com/invite/abc123",
	}

	comms := New(WithDriver(mock.NewDriver()))

	err := comms.SendInviteEmail(context.Background(), recipient, data)
	require.NoError(t, err)

	msg, ok := comms.config.Driver.(*mock.Driver).LastMessage()
	require.True(t, ok)
	require.NotNil(t, msg)

	require.Contains(t, msg.Subject, data.Actor)
	require.Contains(t, msg.Subject, data.Organization)
	require.Equal(t, []string{recipient.String()}, msg.To)

	require.NotEmpty(t, msg.Text)
	require.Contains(t, msg.Text, data.Actor)
	require.Contains(t, msg.Text, data.ActorEmail)
	require.Contains(t, msg.Text, data.Organization)
	require.Contains(t, msg.Text, data.Message)
	require.Contains(t, msg.Text, data.URL)

	// HTML is optional, but should be rendered (even if same as text)
	require.NotEmpty(t, msg.HTML)
}

func TestComms_SendInviteEmail_InvalidRecipient(t *testing.T) {
	recipient := Recipient{Email: ""}
	data := InviteEmailData{
		Actor:        "Max Mustermann",
		ActorEmail:   "max@example.com",
		Organization: "TestOrg",
		Message:      "Let's build something great together!",
		URL:          "https://kopexa.com/invite/abc123",
	}
	comms := New(WithDriver(mock.NewDriver()))
	err := comms.SendInviteEmail(context.Background(), recipient, data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid recipient")
}
