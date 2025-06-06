// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecipient_NewRecipient(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		firstName string
		lastName  string
		want      Recipient
	}{
		{
			name:      "complete information",
			email:     "test@example.com",
			firstName: "John",
			lastName:  "Doe",
			want: Recipient{
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
			},
		},
		{
			name:      "with spaces",
			email:     " test@example.com ",
			firstName: " John ",
			lastName:  " Doe ",
			want: Recipient{
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
			},
		},
		{
			name:      "only email",
			email:     "test@example.com",
			firstName: "",
			lastName:  "",
			want: Recipient{
				Email:     "test@example.com",
				FirstName: "",
				LastName:  "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRecipient(tt.email, tt.firstName, tt.lastName)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestRecipient_Validate(t *testing.T) {
	tests := []struct {
		name      string
		recipient Recipient
		wantErr   bool
	}{
		{
			name: "valid recipient",
			recipient: Recipient{
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: false,
		},
		{
			name: "empty email",
			recipient: Recipient{
				Email:     "",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: true,
		},
		{
			name: "only email",
			recipient: Recipient{
				Email:     "test@example.com",
				FirstName: "",
				LastName:  "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.recipient.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRecipient_Name(t *testing.T) {
	tests := []struct {
		name      string
		recipient Recipient
		want      string
	}{
		{
			name: "full name",
			recipient: Recipient{
				FirstName: "John",
				LastName:  "Doe",
			},
			want: "John Doe",
		},
		{
			name: "only first name",
			recipient: Recipient{
				FirstName: "John",
				LastName:  "",
			},
			want: "John",
		},
		{
			name: "only last name",
			recipient: Recipient{
				FirstName: "",
				LastName:  "Doe",
			},
			want: "Doe",
		},
		{
			name: "no name",
			recipient: Recipient{
				FirstName: "",
				LastName:  "",
			},
			want: "",
		},
		{
			name: "with spaces",
			recipient: Recipient{
				FirstName: " John ",
				LastName:  " Doe ",
			},
			want: "John Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.recipient.Name()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestRecipient_String(t *testing.T) {
	tests := []struct {
		name      string
		recipient Recipient
		want      string
	}{
		{
			name: "with name",
			recipient: Recipient{
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
			},
			want: "John Doe <test@example.com>",
		},
		{
			name: "only email",
			recipient: Recipient{
				Email:     "test@example.com",
				FirstName: "",
				LastName:  "",
			},
			want: "test@example.com",
		},
		{
			name: "with spaces",
			recipient: Recipient{
				Email:     " test@example.com ",
				FirstName: " John ",
				LastName:  " Doe ",
			},
			want: "John Doe <test@example.com>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.recipient.String()
			require.Equal(t, tt.want, got)
		})
	}
}
