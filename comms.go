// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

// Package comms provides a flexible and extensible email communication system.
// It supports template-based emails, multiple email drivers, and various
// customization options.
package comms

// Comms is the main struct that provides email communication functionality.
// It holds the configuration and provides methods for sending different types
// of emails.
type Comms struct {
	config Config
}

// CommonData represents the base data structure that is available in all email templates.
// It contains common fields that are typically needed across different email types.
type CommonData struct {
	// Subject is the email subject line
	Subject string `json:"subject"`
	// Recipient contains information about the email recipient
	Recipient Recipient `json:"recipient"`
}

// New creates a new Comms instance with the given options.
// The options allow for customization of the email system's behavior.
//
// Example:
//
//	comms := comms.New(
//	  comms.WithDriver(driver),
//	  comms.WithTemplatesDir("templates"),
//	)
func New(opts ...Option) *Comms {
	comms := &Comms{
		config: Config{},
	}

	for _, opt := range opts {
		opt(&comms.config)
	}

	return comms
}
