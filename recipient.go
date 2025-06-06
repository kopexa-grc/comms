// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"errors"
	"fmt"
	"strings"
)

// Recipient represents an email recipient with their contact information.
// It provides methods for validation and formatting of recipient data.
type Recipient struct {
	// Email is the recipient's email address
	Email string `json:"email"`
	// FirstName is the recipient's first name
	FirstName string `json:"first_name"`
	// LastName is the recipient's last name
	LastName string `json:"last_name"`
}

var ErrEmailRequired = errors.New("email is required")

// NewRecipient creates a new Recipient instance with the given information.
// All fields are optional except email, which will be validated when needed.
func NewRecipient(email, firstName, lastName string) Recipient {
	return Recipient{
		Email:     strings.TrimSpace(email),
		FirstName: strings.TrimSpace(firstName),
		LastName:  strings.TrimSpace(lastName),
	}
}

// Validate checks if the recipient's data is valid.
// It ensures that the email address is present.
//
// Returns an error if:
// - The email address is empty
func (r *Recipient) Validate() error {
	if r.Email == "" {
		return fmt.Errorf("%w", ErrEmailRequired)
	}

	return nil
}

// Name returns the recipient's full name.
// If both first and last name are empty, it returns an empty string.
func (r *Recipient) Name() string {
	firstName := strings.TrimSpace(r.FirstName)
	lastName := strings.TrimSpace(r.LastName)

	if firstName == "" && lastName == "" {
		return ""
	}

	return strings.TrimSpace(fmt.Sprintf("%s %s", firstName, lastName))
}

// String returns a formatted string representation of the recipient.
// Format: "First Last <email@example.com>" or just "email@example.com" if no name is available.
func (r Recipient) String() string {
	if r.Name() == "" {
		return strings.TrimSpace(r.Email)
	}

	return fmt.Sprintf("%s <%s>", r.Name(), strings.TrimSpace(r.Email))
}
