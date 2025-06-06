// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

// Package driver provides an interface for sending emails through different email service providers.
//
// # Overview
//
// The driver package defines a common interface for email delivery services. It abstracts away
// the implementation details of different email providers, allowing for easy switching between
// services like SMTP, SendGrid, Amazon SES, etc.
//
// # Usage
//
// To use the driver package, first implement the Driver interface for your chosen email service:
//
//	type SMTPDriver struct {
//		host     string
//		port     int
//		username string
//		password string
//	}
//
//	func (d *SMTPDriver) Send(ctx context.Context, message Message) error {
//		// Implementation
//	}
//
// Then use it in your application:
//
//	driver := &SMTPDriver{
//		host:     "smtp.example.com",
//		port:     587,
//		username: "user",
//		password: "pass",
//	}
//
//	err := driver.Send(ctx, Message{
//		From:    "sender@example.com",
//		To:      []string{"recipient@example.com"},
//		Subject: "Hello",
//		HTML:    "<p>Hello World</p>",
//		Text:    "Hello World",
//	})
//
// # Message Structure
//
// The Message struct represents an email message with the following fields:
//
//   - From:    Sender's email address
//   - To:      List of recipient email addresses
//   - Subject: Email subject line
//   - Bcc:     List of blind carbon copy recipients
//   - Cc:      List of carbon copy recipients
//   - ReplyTo: Reply-to email address
//   - HTML:    HTML version of the email body
//   - Text:    Plain text version of the email body
//
// # Context Support
//
// The Send method accepts a context.Context parameter, allowing for:
//   - Timeout control
//   - Cancellation
//   - Request tracing
//   - Deadline management
//
// # Error Handling
//
// The Send method returns an error if the email could not be sent. Common errors include:
//   - Invalid email addresses
//   - Authentication failures
//   - Network issues
//   - Rate limiting
//   - Service unavailability
//
// # Thread Safety
//
// Driver implementations should be safe for concurrent use by multiple goroutines.
package driver
