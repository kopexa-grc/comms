// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package mock

import (
	"context"
	"errors"
	"testing"

	"github.com/kopexa-grc/comms/driver"
	"github.com/stretchr/testify/require"
)

var ErrCallback = errors.New("callback error")

func TestMockDriver_Send(t *testing.T) {
	tests := []struct {
		name    string
		message driver.Message
		onSend  func(_ driver.Message) error
		wantErr bool
	}{
		{
			name: "successful send",
			message: driver.Message{
				From:    "test@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				HTML:    "<p>Test HTML</p>",
				Text:    "Test Text",
			},
			wantErr: false,
		},
		{
			name: "send with callback error",
			message: driver.Message{
				From:    "test@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
			},
			//revive:disable:unused-parameter
			onSend: func(message driver.Message) error {
				return ErrCallback
			},
			//revive:enable:unused-parameter
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDriver()
			d.OnSend = tt.onSend

			err := d.Send(context.Background(), tt.message)

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, 0, d.MessageCount())
			} else {
				require.NoError(t, err)
				require.Equal(t, 1, d.MessageCount())

				lastMessage, ok := d.LastMessage()
				require.True(t, ok)
				require.Equal(t, tt.message, lastMessage)
			}
		})
	}
}

func TestMockDriver_Messages(t *testing.T) {
	d := NewDriver()

	// Send multiple messages
	messages := []driver.Message{
		{
			From:    "test1@example.com",
			To:      []string{"recipient1@example.com"},
			Subject: "Test 1",
		},
		{
			From:    "test2@example.com",
			To:      []string{"recipient2@example.com"},
			Subject: "Test 2",
		},
	}

	for _, msg := range messages {
		err := d.Send(context.Background(), msg)
		require.NoError(t, err)
	}

	// Test Messages()
	storedMessages := d.Messages()
	require.Equal(t, len(messages), len(storedMessages))

	for i, msg := range messages {
		require.Equal(t, msg, storedMessages[i])
	}

	// Test Clear()
	d.Clear()
	require.Equal(t, 0, d.MessageCount())
	require.Empty(t, d.Messages())

	// Test LastMessage() with empty messages
	_, ok := d.LastMessage()
	require.False(t, ok)
}

func TestMockDriver_Concurrent(t *testing.T) {
	mock := NewDriver()
	message := driver.Message{
		From:    "test@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
	}

	// Send multiple messages concurrently
	const goroutines = 10

	done := make(chan struct{})

	for i := 0; i < goroutines; i++ {
		go func() {
			err := mock.Send(context.Background(), message)
			require.NoError(t, err)
			done <- struct{}{}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Verify all messages were stored
	require.Equal(t, goroutines, mock.MessageCount())
	messages := mock.Messages()
	require.Equal(t, goroutines, len(messages))

	for _, msg := range messages {
		require.Equal(t, message, msg)
	}
}
