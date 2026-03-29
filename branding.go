// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

// Default branding values used when Branding fields are zero-valued.
const (
	DefaultBrandName       = "Kopexa"
	DefaultPrimaryColor    = "#2563eb"
	DefaultBackgroundColor = "#f1f5f9"
	DefaultTextColor       = "#0f172a"
	DefaultButtonColor     = "#2563eb"
	DefaultButtonTextColor = "#ffffff"
	DefaultLinkColor       = "#2563eb"
	DefaultFontFamily      = "Helvetica, Arial, sans-serif"
	DefaultCompanyName     = "Kopexa GmbH"
	DefaultCompanyAddress  = "Schauenburgerstr. 116\n24118 Kiel\nGermany"
	DefaultSupportEmail    = "support@kopexa.com"
)

// Branding configures the visual appearance of the email shell.
// Zero-valued fields fall back to Kopexa defaults.
type Branding struct {
	// BrandName is displayed in the email header and footer.
	// Default: "Kopexa"
	BrandName string

	// LogoURL is the URL to the brand logo image shown in the email header.
	// If empty, BrandName is displayed as styled text instead.
	LogoURL string

	// PrimaryColor is the primary accent color (hex) used for the header.
	// Default: "#2563eb"
	PrimaryColor string

	// BackgroundColor is the email body background color (hex).
	// Default: "#f1f5f9"
	BackgroundColor string

	// TextColor is the main body text color (hex).
	// Default: "#0f172a"
	TextColor string

	// ButtonColor is the CTA button background color (hex).
	// Available in body templates via {{.Branding.ButtonColor}}.
	// Default: "#2563eb"
	ButtonColor string

	// ButtonTextColor is the CTA button text color (hex).
	// Available in body templates via {{.Branding.ButtonTextColor}}.
	// Default: "#ffffff"
	ButtonTextColor string

	// LinkColor is the link color (hex) used in the footer.
	// Default: "#2563eb"
	LinkColor string

	// FontFamily is the CSS font-family for the email.
	// Default: "Helvetica, Arial, sans-serif"
	FontFamily string

	// CompanyName is the legal entity name shown in the footer.
	// Default: "Kopexa GmbH"
	CompanyName string

	// CompanyAddress is the multiline company address shown in the footer.
	// Default: "Schauenburgerstr. 116\n24118 Kiel\nGermany"
	CompanyAddress string

	// SupportEmail is the support contact email shown in the footer.
	// Default: "support@kopexa.com"
	SupportEmail string
}

// resolveBranding fills zero-valued fields in the given Branding with defaults.
// If b is nil, a Branding with all defaults is returned.
func resolveBranding(b *Branding) Branding {
	if b == nil {
		return Branding{
			BrandName:       DefaultBrandName,
			PrimaryColor:    DefaultPrimaryColor,
			BackgroundColor: DefaultBackgroundColor,
			TextColor:       DefaultTextColor,
			ButtonColor:     DefaultButtonColor,
			ButtonTextColor: DefaultButtonTextColor,
			LinkColor:       DefaultLinkColor,
			FontFamily:      DefaultFontFamily,
			CompanyName:     DefaultCompanyName,
			CompanyAddress:  DefaultCompanyAddress,
			SupportEmail:    DefaultSupportEmail,
		}
	}

	resolved := *b

	if resolved.BrandName == "" {
		resolved.BrandName = DefaultBrandName
	}

	if resolved.PrimaryColor == "" {
		resolved.PrimaryColor = DefaultPrimaryColor
	}

	if resolved.BackgroundColor == "" {
		resolved.BackgroundColor = DefaultBackgroundColor
	}

	if resolved.TextColor == "" {
		resolved.TextColor = DefaultTextColor
	}

	if resolved.ButtonColor == "" {
		resolved.ButtonColor = DefaultButtonColor
	}

	if resolved.ButtonTextColor == "" {
		resolved.ButtonTextColor = DefaultButtonTextColor
	}

	if resolved.LinkColor == "" {
		resolved.LinkColor = DefaultLinkColor
	}

	if resolved.FontFamily == "" {
		resolved.FontFamily = DefaultFontFamily
	}

	if resolved.CompanyName == "" {
		resolved.CompanyName = DefaultCompanyName
	}

	if resolved.CompanyAddress == "" {
		resolved.CompanyAddress = DefaultCompanyAddress
	}

	if resolved.SupportEmail == "" {
		resolved.SupportEmail = DefaultSupportEmail
	}

	return resolved
}
