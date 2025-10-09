// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package resend_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/kopexa-grc/comms/v2/driver"
	impl "github.com/kopexa-grc/comms/v2/driver/resend"
	"github.com/resend/resend-go/v2"
	"github.com/stretchr/testify/require"
)

func mockClient(t *testing.T, apiKey string, success bool, validateRequest func(r *http.Request)) (*resend.Client, *httptest.Server) {
	var ts *httptest.Server

	if success {
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if validateRequest != nil {
				validateRequest(r)
			}

			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"id": "sent"}`))
			require.NoError(t, err)
		}))
	} else {
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if validateRequest != nil {
				validateRequest(r)
			}

			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(`{"error": "failed to send email"}`))
			require.NoError(t, err)
		}))
	}

	mockClient := resend.NewClient(apiKey)
	baseURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	mockClient.BaseURL = baseURL

	return mockClient, ts
}

func TestResendDriver_Send(t *testing.T) {
	tests := []struct {
		name            string
		message         driver.Message
		success         bool
		wantErr         bool
		validateRequest func(r *http.Request)
	}{
		{
			name: "successful email send",
			message: driver.Message{
				From:    "test@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				HTML:    "<p>Test HTML</p>",
				Text:    "Test Text",
			},
			success: true,
			wantErr: false,
			validateRequest: func(r *http.Request) {
				require.Equal(t, "POST", r.Method)
				require.Equal(t, "/emails", r.URL.Path)

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var request resend.SendEmailRequest
				err = json.Unmarshal(body, &request)
				require.NoError(t, err)

				require.Equal(t, "test@example.com", request.From)
				require.Equal(t, []string{"recipient@example.com"}, request.To)
				require.Equal(t, "Test Subject", request.Subject)
				require.Equal(t, "<p>Test HTML</p>", request.Html)
				require.Equal(t, "Test Text", request.Text)
			},
		},
		{
			name: "email with tags",
			message: driver.Message{
				From:    "test@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				HTML:    "<p>Test HTML</p>",
				Text:    "Test Text",
				Tags: []driver.Tag{
					{Name: "category", Value: "test"},
				},
			},
			success: true,
			wantErr: false,
			validateRequest: func(r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var request resend.SendEmailRequest
				err = json.Unmarshal(body, &request)
				require.NoError(t, err)

				require.Len(t, request.Tags, 1)
				require.Equal(t, "category", request.Tags[0].Name)
				require.Equal(t, "test", request.Tags[0].Value)
			},
		},
		{
			name: "email with attachments",
			message: driver.Message{
				From:    "test@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				HTML:    "<p>Test HTML</p>",
				Text:    "Test Text",
				Attachments: []driver.Attachment{
					{
						Filename:    "test.txt",
						Content:     []byte("test content"),
						ContentType: "text/plain",
					},
				},
			},
			success: true,
			wantErr: false,
			validateRequest: func(r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var request resend.SendEmailRequest
				err = json.Unmarshal(body, &request)
				require.NoError(t, err)

				require.Len(t, request.Attachments, 1)
				require.Equal(t, "test.txt", request.Attachments[0].Filename)
				require.Equal(t, []byte("test content"), request.Attachments[0].Content)
				if request.Attachments[0].ContentType != "" {
					require.Equal(t, "text/plain", request.Attachments[0].ContentType)
				}
			},
		},
		{
			name: "email with default content type",
			message: driver.Message{
				From:    "test@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				HTML:    "<p>Test HTML</p>",
				Text:    "Test Text",
				Attachments: []driver.Attachment{
					{
						Filename: "test.bin",
						Content:  []byte("test content"),
					},
				},
			},
			success: true,
			wantErr: false,
			validateRequest: func(r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var request resend.SendEmailRequest
				err = json.Unmarshal(body, &request)
				require.NoError(t, err)

				require.Len(t, request.Attachments, 1)
				require.Equal(t, "test.bin", request.Attachments[0].Filename)
				require.Equal(t, []byte("test content"), request.Attachments[0].Content)
				if request.Attachments[0].ContentType != "" {
					require.Equal(t, "application/octet-stream", request.Attachments[0].ContentType)
				}
			},
		},
		{
			name: "email with headers",
			message: driver.Message{
				From:    "test@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				HTML:    "<p>Test HTML</p>",
				Text:    "Test Text",
				Headers: map[string]string{
					"X-Mailer": "Test Mailer",
				},
			},
			success: true,
			wantErr: false,
			validateRequest: func(r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var request resend.SendEmailRequest
				err = json.Unmarshal(body, &request)
				require.NoError(t, err)

				require.Equal(t, "Test Mailer", request.Headers["X-Mailer"])
			},
		},
		{
			name: "failed email send",
			message: driver.Message{
				From:    "test@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				HTML:    "<p>Test HTML</p>",
				Text:    "Test Text",
			},
			success: false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, server := mockClient(t, "test-api-key", tt.success, tt.validateRequest)
			defer server.Close()

			driver := impl.New("", impl.WithClient(client))
			err := driver.Send(context.Background(), tt.message)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestResendDriver_Validate(t *testing.T) {
	tests := []struct {
		name    string
		message driver.Message
		wantErr bool
	}{
		{
			name: "valid message",
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
			name: "invalid from address",
			message: driver.Message{
				From:    "invalid-email",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				HTML:    "<p>Test HTML</p>",
				Text:    "Test Text",
			},
			wantErr: true,
		},
		{
			name: "missing to address",
			message: driver.Message{
				From:    "test@example.com",
				To:      []string{},
				Subject: "Test Subject",
				HTML:    "<p>Test HTML</p>",
				Text:    "Test Text",
			},
			wantErr: true,
		},
		{
			name: "invalid to address",
			message: driver.Message{
				From:    "test@example.com",
				To:      []string{"invalid-email"},
				Subject: "Test Subject",
				HTML:    "<p>Test HTML</p>",
				Text:    "Test Text",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, server := mockClient(t, "test-api-key", true, nil)
			defer server.Close()

			driver := impl.New("", impl.WithClient(client))
			err := driver.Send(context.Background(), tt.message)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
