// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package driver

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
)

// Driver defines the interface for email delivery services.
// Implementations of this interface should handle the actual sending of emails
// through various email service providers (SMTP, SendGrid, Amazon SES, etc.).
type Driver interface {
	// Send delivers an email message using the configured email service.
	// The context can be used to control timeouts, cancellation, and deadlines.
	// Returns an error if the email could not be sent.
	Send(ctx context.Context, message Message) error
}

// Message represents an email message with all its components.
// Both HTML and Text fields should be provided for maximum email client compatibility.
type Message struct {
	// From is the sender's email address.
	From string `json:"from"`

	// To is the list of primary recipient email addresses.
	To []string `json:"to"`

	// Subject is the email subject line.
	Subject string `json:"subject"`

	// Bcc is the list of blind carbon copy recipient email addresses.
	// These recipients will receive the email but won't be visible to other recipients.
	Bcc []string `json:"bcc"`

	// Cc is the list of carbon copy recipient email addresses.
	// These recipients will receive the email and be visible to all other recipients.
	Cc []string `json:"cc"`

	// ReplyTo is the email address that should receive replies to this message.
	// If not set, replies will go to the From address.
	ReplyTo string `json:"reply_to"`

	// HTML is the HTML version of the email body.
	// Should be provided for rich email clients.
	HTML string `json:"html"`

	// Text is the plain text version of the email body.
	// Should be provided for email clients that don't support HTML or for accessibility.
	Text string `json:"text"`

	// Tags can be used for categorization, tracking, or filtering emails.
	Tags []Tag `json:"tags"`

	// Attachments is a list of files to include with the email.
	// Each attachment must have a filename and content.
	Attachments []Attachment `json:"attachments"`

	// Headers contains additional email headers to include.
	// These headers will be added to the email message.
	// Common headers include "X-Mailer", "List-Unsubscribe", etc.
	Headers map[string]string `json:"headers,omitempty"`
}

// Tag represents a key-value pair that can be attached to an email.
// Tags are useful for categorizing and tracking emails in email service providers.
type Tag struct {
	// Name is the tag identifier.
	Name string `json:"name"`

	// Value is the tag's value.
	Value string `json:"value"`
}

// Attachment represents a file to be attached to an email.
type Attachment struct {
	// Filename is the name of the attachment file.
	// This is the name that will be shown to the recipient.
	Filename string

	// Content is the binary content of the attachment.
	// This field is used when the attachment content is already in memory.
	Content []byte

	// ContentType is the MIME type of the file.
	// If not specified, it will be determined from the file extension.
	ContentType string

	// FilePath is the path to the attachment file.
	// This field is used when the attachment should be loaded from disk.
	FilePath string
}

var ErrToAddressRequired = errors.New("to address is required")

// Validate performs basic validation of the email message.
// It checks:
//   - From address is a valid email address
//   - At least one To address is provided
//   - All To addresses are valid email addresses
func (m *Message) Validate() error {
	if _, err := mail.ParseAddress(m.From); err != nil {
		return fmt.Errorf("invalid from address: %w", err)
	}

	if len(m.To) == 0 {
		return fmt.Errorf("%w", ErrToAddressRequired)
	}

	for _, to := range m.To {
		if _, err := mail.ParseAddress(to); err != nil {
			return fmt.Errorf("invalid to address: %w", err)
		}
	}

	return nil
}
