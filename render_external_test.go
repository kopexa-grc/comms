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

func TestRenderTextTemplate(t *testing.T) {
	t.Run("simple variable substitution", func(t *testing.T) {
		result, err := renderTextTmpl("Hello {{.Name}}", map[string]any{"Name": "Max"})
		require.NoError(t, err)
		require.Equal(t, "Hello Max", result)
	})

	t.Run("sprig function", func(t *testing.T) {
		result, err := renderTextTmpl("{{.Name | upper}}", map[string]any{"Name": "max"})
		require.NoError(t, err)
		require.Equal(t, "MAX", result)
	})

	t.Run("empty template returns empty string", func(t *testing.T) {
		result, err := renderTextTmpl("", map[string]any{"Name": "Max"})
		require.NoError(t, err)
		require.Empty(t, result)
	})

	t.Run("invalid syntax returns error", func(t *testing.T) {
		_, err := renderTextTmpl("{{.Name", map[string]any{"Name": "Max"})
		require.Error(t, err)
	})

	t.Run("missing variable returns error", func(t *testing.T) {
		_, err := renderTextTmpl("{{.Missing}}", map[string]any{})
		require.Error(t, err)
	})
}

func TestRenderHTMLTmpl(t *testing.T) {
	t.Run("html escapes by default", func(t *testing.T) {
		result, err := renderHTMLTmpl("Hello {{.Name}}", map[string]any{"Name": "<script>alert('xss')</script>"})
		require.NoError(t, err)
		require.NotContains(t, result, "<script>")
		require.Contains(t, result, "&lt;script&gt;")
	})

	t.Run("sprig function", func(t *testing.T) {
		result, err := renderHTMLTmpl("{{.Name | upper}}", map[string]any{"Name": "max"})
		require.NoError(t, err)
		require.Equal(t, "MAX", result)
	})

	t.Run("empty template returns empty string", func(t *testing.T) {
		result, err := renderHTMLTmpl("", map[string]any{"Name": "Max"})
		require.NoError(t, err)
		require.Empty(t, result)
	})

	t.Run("invalid syntax returns error", func(t *testing.T) {
		_, err := renderHTMLTmpl("{{.Name", map[string]any{"Name": "Max"})
		require.Error(t, err)
	})
}

func TestRenderShell(t *testing.T) {
	branding := resolveBranding(nil)

	t.Run("includes DOCTYPE", func(t *testing.T) {
		result, err := renderShell("Hello", "", branding)
		require.NoError(t, err)
		require.Contains(t, result, "<!DOCTYPE html")
	})

	t.Run("includes body content", func(t *testing.T) {
		result, err := renderShell("<h1>Test Content</h1>", "", branding)
		require.NoError(t, err)
		require.Contains(t, result, "<h1>Test Content</h1>")
	})

	t.Run("applies branding background color", func(t *testing.T) {
		result, err := renderShell("content", "", branding)
		require.NoError(t, err)
		require.Contains(t, result, DefaultBackgroundColor)
	})

	t.Run("applies branding text color", func(t *testing.T) {
		result, err := renderShell("content", "", branding)
		require.NoError(t, err)
		require.Contains(t, result, DefaultTextColor)
	})

	t.Run("applies branding font family", func(t *testing.T) {
		result, err := renderShell("content", "", branding)
		require.NoError(t, err)
		require.Contains(t, result, DefaultFontFamily)
	})

	t.Run("shows brand name when no logo URL", func(t *testing.T) {
		result, err := renderShell("content", "", branding)
		require.NoError(t, err)
		require.Contains(t, result, DefaultBrandName)
	})

	t.Run("shows logo img when logo URL set", func(t *testing.T) {
		b := resolveBranding(&Branding{
			LogoURL: "https://example.com/logo.png",
		})
		result, err := renderShell("content", "", b)
		require.NoError(t, err)
		require.Contains(t, result, `src="https://example.com/logo.png"`)
	})

	t.Run("includes preheader when set", func(t *testing.T) {
		result, err := renderShell("content", "Preview text here", branding)
		require.NoError(t, err)
		require.Contains(t, result, "Preview text here")
	})

	t.Run("includes footer with company name", func(t *testing.T) {
		result, err := renderShell("content", "", branding)
		require.NoError(t, err)
		require.Contains(t, result, DefaultCompanyName)
	})

	t.Run("includes footer with support email", func(t *testing.T) {
		result, err := renderShell("content", "", branding)
		require.NoError(t, err)
		require.Contains(t, result, DefaultSupportEmail)
	})

	t.Run("custom branding colors applied", func(t *testing.T) {
		b := resolveBranding(&Branding{
			BackgroundColor: "#111111",
			TextColor:       "#eeeeee",
			PrimaryColor:    "#ff0000",
		})
		result, err := renderShell("content", "", b)
		require.NoError(t, err)
		require.Contains(t, result, "#111111")
		require.Contains(t, result, "#eeeeee")
		require.Contains(t, result, "#ff0000")
	})
}
