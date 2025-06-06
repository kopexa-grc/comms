// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"testing"

	"github.com/kopexa-grc/comms/driver/mock"
	"github.com/stretchr/testify/require"
)

func TestComms_SendInviteAcceptedEmail(t *testing.T) {
	recipient := Recipient{
		Email:     "inviter@example.com",
		FirstName: "Max",
		LastName:  "Mustermann",
	}
	data := InviteAcceptedData{
		InviteeName:  "Alice Example",
		Organization: "Acme GmbH",
	}

	comms := New(WithDriver(mock.NewDriver()))

	err := comms.SendInviteAcceptedEmail(context.Background(), recipient, data)
	require.NoError(t, err)

	msg, ok := comms.config.Driver.(*mock.Driver).LastMessage()
	require.True(t, ok)
	require.NotNil(t, msg)

	require.Contains(t, msg.Subject, data.InviteeName)
	require.Contains(t, msg.Subject, data.Organization)
	require.Equal(t, []string{recipient.String()}, msg.To)

	require.NotEmpty(t, msg.Text)
	require.Contains(t, msg.Text, data.InviteeName)
	require.Contains(t, msg.Text, data.Organization)
	require.Contains(t, msg.Text, "Good news")

	require.NotEmpty(t, msg.HTML)
	require.Contains(t, msg.HTML, data.InviteeName)
	require.Contains(t, msg.HTML, data.Organization)
	require.Contains(t, msg.HTML, "<!DOCTYPE html PUBLIC")
}

func TestComms_SendInviteAcceptedEmail_InvalidRecipient(t *testing.T) {
	recipient := Recipient{Email: ""}
	data := InviteAcceptedData{
		InviteeName:  "Alice Example",
		Organization: "Acme GmbH",
	}
	comms := New(WithDriver(mock.NewDriver()))
	err := comms.SendInviteAcceptedEmail(context.Background(), recipient, data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid recipient")
} 