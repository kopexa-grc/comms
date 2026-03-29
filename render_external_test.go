// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergeData(t *testing.T) {
	t.Run("nil defaults and nil data", func(t *testing.T) {
		branding := resolveBranding(nil)
		result := mergeData(nil, nil, branding)

		require.NotNil(t, result)
		require.Equal(t, branding, result["Branding"])
	})

	t.Run("defaults only", func(t *testing.T) {
		defaults := map[string]any{"AppURL": "https://app.example.com"}
		branding := resolveBranding(nil)
		result := mergeData(defaults, nil, branding)

		require.Equal(t, "https://app.example.com", result["AppURL"])
		require.Equal(t, branding, result["Branding"])
	})

	t.Run("data only", func(t *testing.T) {
		data := map[string]any{"Name": "Max"}
		branding := resolveBranding(nil)
		result := mergeData(nil, data, branding)

		require.Equal(t, "Max", result["Name"])
		require.Equal(t, branding, result["Branding"])
	})

	t.Run("data wins over defaults", func(t *testing.T) {
		defaults := map[string]any{
			"AppURL": "https://default.example.com",
			"Footer": "Default footer",
		}
		data := map[string]any{
			"AppURL": "https://custom.example.com",
			"Name":   "Max",
		}
		branding := resolveBranding(nil)
		result := mergeData(defaults, data, branding)

		require.Equal(t, "https://custom.example.com", result["AppURL"])
		require.Equal(t, "Default footer", result["Footer"])
		require.Equal(t, "Max", result["Name"])
	})

	t.Run("Branding key cannot be overridden by data", func(t *testing.T) {
		data := map[string]any{
			"Branding": "should be ignored",
		}
		branding := resolveBranding(nil)
		result := mergeData(nil, data, branding)

		require.Equal(t, branding, result["Branding"])
	})
}
