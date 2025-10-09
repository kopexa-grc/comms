// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"errors"
	"fmt"

	"github.com/kopexa-grc/comms/v2/driver"
)

var (
	ErrDriverRequired = errors.New("driver is required")
	ErrFromRequired   = errors.New("from is required")
)

type Config struct {
	Driver driver.Driver
	From   string
}

func (c *Config) Validate() error {
	if c.Driver == nil {
		return fmt.Errorf("%w", ErrDriverRequired)
	}

	if c.From == "" {
		return fmt.Errorf("%w", ErrFromRequired)
	}

	return nil
}

type Option func(*Config)

func WithDriver(driver driver.Driver) Option {
	return func(c *Config) {
		c.Driver = driver
	}
}

func WithFrom(from string) Option {
	return func(c *Config) {
		c.From = from
	}
}
