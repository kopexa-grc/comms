// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package mock

import (
	"context"
	"sync"

	"github.com/kopexa-grc/comms/v2/driver"
)

// Driver is a mock implementation of the driver.Driver interface
// for testing purposes. It stores all sent messages and allows
// inspection of the sent messages.
type Driver struct {
	mu       sync.RWMutex
	messages []driver.Message
	// OnSend is an optional callback that is called before sending
	OnSend func(message driver.Message) error
}

// NewDriver creates a new instance of Driver
func NewDriver() *Driver {
	return &Driver{
		messages: make([]driver.Message, 0),
	}
}

// Send implements the driver.Driver interface
func (m *Driver) Send(_ context.Context, message driver.Message) error {
	if m.OnSend != nil {
		if err := m.OnSend(message); err != nil {
			return err
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = append(m.messages, message)

	return nil
}

// Messages returns all sent messages
func (m *Driver) Messages() []driver.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy of messages to avoid race conditions
	messages := make([]driver.Message, len(m.messages))
	copy(messages, m.messages)

	return messages
}

// Clear removes all stored messages
func (m *Driver) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = make([]driver.Message, 0)
}

// LastMessage returns the last sent message
func (m *Driver) LastMessage() (driver.Message, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.messages) == 0 {
		return driver.Message{}, false
	}

	return m.messages[len(m.messages)-1], true
}

// MessageCount returns the number of sent messages
func (m *Driver) MessageCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.messages)
}
