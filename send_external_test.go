// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"testing"

	"github.com/kopexa-grc/comms/v2/driver"
	"github.com/kopexa-grc/comms/v2/driver/mock"
	"github.com/stretchr/testify/require"
)

var errDriverSend = driver.ErrToAddressRequired // sentinel reused for OnSend callback tests

func TestComms_SendRendered(t *testing.T) {
	t.Run("sends rendered email via driver", func(t *testing.T) {
		d := mock.NewDriver()
		c := New(WithDriver(d), WithFrom("noreply@example.com"))
		recipient := Recipient{Email: "user@example.com", FirstName: "Max"}

		rendered := &RenderedEmail{
			Subject: "Test Subject",
			HTML:    "<p>HTML body</p>",
			Text:    "Text body",
		}

		err := c.SendRendered(context.Background(), recipient, rendered)
		require.NoError(t, err)

		msg, ok := d.LastMessage()
		require.True(t, ok)
		require.Equal(t, "noreply@example.com", msg.From)
		require.Equal(t, []string{recipient.String()}, msg.To)
		require.Equal(t, "Test Subject", msg.Subject)
		require.Equal(t, "<p>HTML body</p>", msg.HTML)
		require.Equal(t, "Text body", msg.Text)
	})

	t.Run("nil rendered returns error", func(t *testing.T) {
		d := mock.NewDriver()
		c := New(WithDriver(d), WithFrom("noreply@example.com"))
		recipient := Recipient{Email: "user@example.com"}

		err := c.SendRendered(context.Background(), recipient, nil)
		require.Error(t, err)
	})

	t.Run("empty subject returns error", func(t *testing.T) {
		d := mock.NewDriver()
		c := New(WithDriver(d), WithFrom("noreply@example.com"))
		recipient := Recipient{Email: "user@example.com"}

		rendered := &RenderedEmail{
			HTML: "<p>body</p>",
		}

		err := c.SendRendered(context.Background(), recipient, rendered)
		require.Error(t, err)
	})

	t.Run("empty body returns error", func(t *testing.T) {
		d := mock.NewDriver()
		c := New(WithDriver(d), WithFrom("noreply@example.com"))
		recipient := Recipient{Email: "user@example.com"}

		rendered := &RenderedEmail{
			Subject: "Subject",
		}

		err := c.SendRendered(context.Background(), recipient, rendered)
		require.Error(t, err)
	})

	t.Run("invalid recipient returns error", func(t *testing.T) {
		d := mock.NewDriver()
		c := New(WithDriver(d), WithFrom("noreply@example.com"))
		recipient := Recipient{Email: ""}

		rendered := &RenderedEmail{
			Subject: "Subject",
			HTML:    "<p>body</p>",
		}

		err := c.SendRendered(context.Background(), recipient, rendered)
		require.Error(t, err)
	})

	t.Run("driver error propagated", func(t *testing.T) {
		d := mock.NewDriver()
		d.OnSend = func(_ driver.Message) error {
			return errDriverSend
		}
		c := New(WithDriver(d), WithFrom("noreply@example.com"))
		recipient := Recipient{Email: "user@example.com"}

		rendered := &RenderedEmail{
			Subject: "Subject",
			HTML:    "<p>body</p>",
		}

		err := c.SendRendered(context.Background(), recipient, rendered)
		require.Error(t, err)
		require.ErrorIs(t, err, errDriverSend)
	})
}

func TestComms_SendFromTemplate(t *testing.T) {
	t.Run("renders and sends in one call", func(t *testing.T) {
		d := mock.NewDriver()
		c := New(WithDriver(d), WithFrom("noreply@example.com"))
		recipient := Recipient{Email: "user@example.com", FirstName: "Max", LastName: "Mustermann"}

		tmpl := ExternalTemplate{
			SubjectTemplate: "Hello {{.Name}}",
			BodyTemplate:    "<p>Welcome, {{.Name}}!</p>",
			TextTemplate:    "Welcome, {{.Name}}!",
		}
		data := map[string]any{"Name": "Max"}

		err := c.SendFromTemplate(context.Background(), recipient, tmpl, nil, data)
		require.NoError(t, err)

		msg, ok := d.LastMessage()
		require.True(t, ok)
		require.Equal(t, "Hello Max", msg.Subject)
		require.Equal(t, []string{recipient.String()}, msg.To)
		require.Contains(t, msg.HTML, "Welcome, Max!")
		require.Contains(t, msg.HTML, "<!DOCTYPE html")
		require.Equal(t, "Welcome, Max!", msg.Text)
	})

	t.Run("render error propagated", func(t *testing.T) {
		d := mock.NewDriver()
		c := New(WithDriver(d), WithFrom("noreply@example.com"))
		recipient := Recipient{Email: "user@example.com"}

		tmpl := ExternalTemplate{} // no content — validation error

		err := c.SendFromTemplate(context.Background(), recipient, tmpl, nil, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrTemplateContentRequired)
	})

	t.Run("with custom branding", func(t *testing.T) {
		d := mock.NewDriver()
		c := New(WithDriver(d), WithFrom("noreply@example.com"))
		recipient := Recipient{Email: "user@example.com"}

		tmpl := ExternalTemplate{
			SubjectTemplate: "Test",
			BodyTemplate:    "<p>Content</p>",
		}
		branding := &Branding{
			BrandName:    "Acme",
			PrimaryColor: "#ff0000",
		}

		err := c.SendFromTemplate(context.Background(), recipient, tmpl, branding, nil)
		require.NoError(t, err)

		msg, ok := d.LastMessage()
		require.True(t, ok)
		require.Contains(t, msg.HTML, "Acme")
		require.Contains(t, msg.HTML, "#ff0000")
	})
}
