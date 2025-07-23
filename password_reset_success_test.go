// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"testing"

	"github.com/kopexa-grc/comms/driver/mock"
	"github.com/stretchr/testify/require"
)

func TestComms_SendPasswordResetSuccessEmail(t *testing.T) {
	recipient := Recipient{
		Email:     "test@example.com",
		FirstName: "Max",
		LastName:  "Mustermann",
	}

	comms := New(WithDriver(mock.NewDriver()))

	err := comms.SendPasswordResetSuccessEmail(context.Background(), recipient)
	require.NoError(t, err)

	msg, ok := comms.config.Driver.(*mock.Driver).LastMessage()
	require.True(t, ok)
	require.NotNil(t, msg)

	require.Equal(t, "Your password has been successfully reset", msg.Subject)
	require.Equal(t, []string{recipient.String()}, msg.To)

	require.NotEmpty(t, msg.Text)
	require.Contains(t, msg.Text, recipient.Name())

	require.NotEmpty(t, msg.HTML)
	require.Contains(t, msg.HTML, recipient.Name())
	require.Contains(t, msg.HTML, "<!DOCTYPE html PUBLIC")
}

func TestComms_SendPasswordResetSuccessEmail_InvalidRecipient(t *testing.T) {
	recipient := Recipient{Email: ""}
	comms := New(WithDriver(mock.NewDriver()))
	err := comms.SendPasswordResetSuccessEmail(context.Background(), recipient)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid recipient")
}
