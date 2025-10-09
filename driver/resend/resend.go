// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package resend

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"strings"

	"github.com/kopexa-grc/comms/v2/driver"
	"github.com/resend/resend-go/v2"
)

// Ensure Resend implements the driver.Driver interface at compile time
var _ driver.Driver = (*Resend)(nil)

// Resend is an implementation of the driver.Driver interface that uses the Resend API
// for sending emails. It provides a simple way to send emails through the Resend service
// with support for all standard email features like HTML/text content, attachments,
// tags, and custom headers.
type Resend struct {
	client *resend.Client
}

// Option is a function that configures a Resend instance
type Option func(*Resend)

// New creates a new Resend driver instance with the given API key and optional
// configuration options. The API key is required for authentication with the Resend service.
//
// Example:
//
//	driver := resend.New("re_123...", resend.WithBaseURL(customURL))
func New(apiKey string, opts ...Option) *Resend {
	s := &Resend{
		client: resend.NewClient(apiKey),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// WithClient returns an Option that sets a custom Resend client instance.
// This is primarily useful for testing or when you need to customize the client behavior.
func WithClient(client *resend.Client) Option {
	return func(r *Resend) {
		r.client = client
	}
}

// WithBaseURL returns an Option that sets a custom base URL for the Resend API.
// This is useful for testing or when you need to use a different API endpoint.
func WithBaseURL(baseURL *url.URL) Option {
	return func(r *Resend) {
		r.client.BaseURL = baseURL
	}
}

// Send implements the driver.Driver interface. It sends an email using the Resend API.
// The method validates the message before sending and handles any errors that occur
// during the sending process.
//
// The method supports:
// - HTML and plain text content
// - Multiple recipients (To, Cc, Bcc)
// - Reply-To addresses
// - Custom headers
// - Tags for categorization
// - File attachments
//
// Returns an error if:
// - The message validation fails
// - The API request fails
// - The API returns an error response
func (r *Resend) Send(ctx context.Context, message driver.Message) error {
	if err := message.Validate(); err != nil {
		return err
	}

	msg := resend.SendEmailRequest{
		From:        message.From,
		To:          message.To,
		Subject:     message.Subject,
		Bcc:         message.Bcc,
		Cc:          message.Cc,
		ReplyTo:     message.ReplyTo,
		Html:        message.HTML,
		Text:        message.Text,
		Tags:        toResendTags(message.Tags),
		Attachments: toResendAttachments(message.Attachments),
		Headers:     maps.Clone(message.Headers),
	}

	_, err := r.client.Emails.SendWithContext(ctx, &msg)
	if err != nil {
		if strings.Contains(err.Error(), "use our testing email address") {
			return fmt.Errorf("resend: %w", err)
		}

		return fmt.Errorf("resend: failed to send email: %w", err)
	}

	return nil
}

// toResendTags converts driver.Tag slices to resend.Tag slices.
// This is an internal helper function that maps our tag structure
// to the Resend API's tag structure.
func toResendTags(tags []driver.Tag) []resend.Tag {
	resendTags := make([]resend.Tag, len(tags))
	for i, tag := range tags {
		resendTags[i] = resend.Tag{
			Name:  tag.Name,
			Value: tag.Value,
		}
	}

	return resendTags
}

// toResendAttachments converts driver.Attachment slices to resend.Attachment slices.
// This is an internal helper function that maps our attachment structure
// to the Resend API's attachment structure.
//
// If no ContentType is specified for an attachment, it defaults to
// "application/octet-stream".
func toResendAttachments(attachments []driver.Attachment) []*resend.Attachment {
	resendAttachments := make([]*resend.Attachment, len(attachments))

	for i, attachment := range attachments {
		contentType := attachment.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		resendAttachments[i] = &resend.Attachment{
			Filename:    attachment.Filename,
			Content:     attachment.Content,
			ContentType: contentType,
			Path:        attachment.FilePath,
		}
	}

	return resendAttachments
}
