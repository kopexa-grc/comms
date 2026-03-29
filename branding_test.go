// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveBranding_Nil(t *testing.T) {
	b := resolveBranding(nil)

	require.Equal(t, DefaultBrandName, b.BrandName)
	require.Equal(t, DefaultPrimaryColor, b.PrimaryColor)
	require.Equal(t, DefaultBackgroundColor, b.BackgroundColor)
	require.Equal(t, DefaultTextColor, b.TextColor)
	require.Equal(t, DefaultButtonColor, b.ButtonColor)
	require.Equal(t, DefaultButtonTextColor, b.ButtonTextColor)
	require.Equal(t, DefaultLinkColor, b.LinkColor)
	require.Equal(t, DefaultFontFamily, b.FontFamily)
	require.Equal(t, DefaultCompanyName, b.CompanyName)
	require.Equal(t, DefaultCompanyAddress, b.CompanyAddress)
	require.Equal(t, DefaultSupportEmail, b.SupportEmail)
	require.Empty(t, b.LogoURL)
}

func TestResolveBranding_Partial(t *testing.T) {
	b := resolveBranding(&Branding{
		BrandName:    "Acme Corp",
		PrimaryColor: "#ff6600",
	})

	require.Equal(t, "Acme Corp", b.BrandName)
	require.Equal(t, "#ff6600", b.PrimaryColor)
	// All other fields should be defaults
	require.Equal(t, DefaultBackgroundColor, b.BackgroundColor)
	require.Equal(t, DefaultTextColor, b.TextColor)
	require.Equal(t, DefaultButtonColor, b.ButtonColor)
	require.Equal(t, DefaultButtonTextColor, b.ButtonTextColor)
	require.Equal(t, DefaultLinkColor, b.LinkColor)
	require.Equal(t, DefaultFontFamily, b.FontFamily)
	require.Equal(t, DefaultCompanyName, b.CompanyName)
	require.Equal(t, DefaultCompanyAddress, b.CompanyAddress)
	require.Equal(t, DefaultSupportEmail, b.SupportEmail)
}

func TestResolveBranding_Full(t *testing.T) {
	input := &Branding{
		BrandName:       "Acme Corp",
		LogoURL:         "https://acme.com/logo.png",
		PrimaryColor:    "#ff6600",
		BackgroundColor: "#000000",
		TextColor:       "#ffffff",
		ButtonColor:     "#cc0000",
		ButtonTextColor: "#00ff00",
		LinkColor:       "#0000ff",
		FontFamily:      "Georgia, serif",
		CompanyName:     "Acme Corp Inc.",
		CompanyAddress:  "123 Main St\nNew York, NY",
		SupportEmail:    "help@acme.com",
	}

	b := resolveBranding(input)

	require.Equal(t, "Acme Corp", b.BrandName)
	require.Equal(t, "https://acme.com/logo.png", b.LogoURL)
	require.Equal(t, "#ff6600", b.PrimaryColor)
	require.Equal(t, "#000000", b.BackgroundColor)
	require.Equal(t, "#ffffff", b.TextColor)
	require.Equal(t, "#cc0000", b.ButtonColor)
	require.Equal(t, "#00ff00", b.ButtonTextColor)
	require.Equal(t, "#0000ff", b.LinkColor)
	require.Equal(t, "Georgia, serif", b.FontFamily)
	require.Equal(t, "Acme Corp Inc.", b.CompanyName)
	require.Equal(t, "123 Main St\nNew York, NY", b.CompanyAddress)
	require.Equal(t, "help@acme.com", b.SupportEmail)
}
