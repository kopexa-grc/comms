# External Templates Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the ability to render and send emails from externally-provided Go template strings with configurable branding, alongside existing embedded templates.

**Architecture:** Two-phase API — `RenderTemplate` (pure package-level function) renders external template strings into a branded HTML email shell, and `SendRendered` / `SendFromTemplate` (methods on `Comms`) send the result via the configured driver. A Go `html/template` shell provides the email chrome (header, footer, branding) using table layout and inline styles for email client compatibility.

**Tech Stack:** Go 1.24, `html/template` + `text/template`, `github.com/Masterminds/sprig/v3`, `github.com/stretchr/testify`

**Spec:** `docs/superpowers/specs/2026-03-29-external-templates-design.md`

---

## File Structure

| File | Responsibility |
|------|---------------|
| `branding.go` | `Branding` type, default constants, `resolveBranding()` |
| `branding_test.go` | Tests for branding default resolution |
| `external_template.go` | `ExternalTemplate` type, `RenderedEmail` type, `Validate()` |
| `external_template_test.go` | Validation tests |
| `shell.go` | HTML email shell template constant |
| `render_external.go` | `RenderTemplate()`, `mergeData()`, `renderShell()`, template rendering helpers |
| `render_external_test.go` | Full render pipeline tests |
| `send_external.go` | `SendRendered()`, `SendFromTemplate()` |
| `send_external_test.go` | Send method tests with mock driver |
| `example_external_test.go` | Testable GoDoc examples |
| `README.md` | New "External Templates" section |

---

### Task 1: Branding Type & Default Resolution

**Files:**
- Create: `branding.go`
- Create: `branding_test.go`

- [ ] **Step 1: Write the failing tests for branding default resolution**

Create `branding_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -run "TestResolveBranding" -v ./...`
Expected: FAIL — `resolveBranding` not defined, `Branding` type not defined

- [ ] **Step 3: Implement Branding type and resolveBranding**

Create `branding.go`:

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -run "TestResolveBranding" -v ./...`
Expected: PASS — all 3 tests green

- [ ] **Step 5: Commit**

```bash
git add branding.go branding_test.go
git commit -m "feat(external): add Branding type with default resolution"
```

---

### Task 2: ExternalTemplate & RenderedEmail Types + Validation

**Files:**
- Create: `external_template.go`
- Create: `external_template_test.go`

- [ ] **Step 1: Write the failing tests for ExternalTemplate validation**

Create `external_template_test.go`:

```go
// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExternalTemplate_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tmpl    ExternalTemplate
		wantErr bool
	}{
		{
			name:    "both empty returns error",
			tmpl:    ExternalTemplate{},
			wantErr: true,
		},
		{
			name: "body only is valid",
			tmpl: ExternalTemplate{
				BodyTemplate: "<h1>Hello {{.Name}}</h1>",
			},
			wantErr: false,
		},
		{
			name: "text only is valid",
			tmpl: ExternalTemplate{
				TextTemplate: "Hello {{.Name}}",
			},
			wantErr: false,
		},
		{
			name: "both set is valid",
			tmpl: ExternalTemplate{
				BodyTemplate: "<h1>Hello {{.Name}}</h1>",
				TextTemplate: "Hello {{.Name}}",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tmpl.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrTemplateContentRequired)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -run "TestExternalTemplate_Validate" -v ./...`
Expected: FAIL — `ExternalTemplate` not defined

- [ ] **Step 3: Implement ExternalTemplate, RenderedEmail, and Validate**

Create `external_template.go`:

```go
// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import "errors"

var ErrTemplateContentRequired = errors.New("at least one of BodyTemplate or TextTemplate is required")

// ExternalTemplate holds template strings loaded from an external source (e.g., database).
// Templates use Go's html/template syntax with Sprig v3 functions available.
//
// The BodyTemplate is rendered inside the email shell — do NOT include
// DOCTYPE, html, or body tags. The shell provides the email chrome
// (header, footer, branding) automatically.
//
// Template variables are passed as map[string]any. Branding values are
// accessible via {{.Branding.ButtonColor}}, {{.Branding.LinkColor}}, etc.
type ExternalTemplate struct {
	// SubjectTemplate is a Go template string for the email subject line.
	// Rendered with text/template (no HTML escaping).
	// Example: "{{.ActorName}} invited you to {{.OrgName}}"
	SubjectTemplate string

	// PreheaderTemplate is a Go template string for the email preview text.
	// This appears in email clients before opening the email.
	// Rendered with text/template (no HTML escaping).
	PreheaderTemplate string

	// BodyTemplate is a Go html/template string for the email body content.
	// Rendered inside the email shell — do NOT include DOCTYPE/html/body tags.
	BodyTemplate string

	// TextTemplate is a Go text/template string for the plain text fallback.
	// If empty, no plain text version is included in the rendered email.
	TextTemplate string

	// Defaults are base-layer variables merged with call-site data.
	// Call-site data takes precedence over defaults.
	Defaults map[string]any
}

// Validate checks that the ExternalTemplate has at least one content template.
// Returns ErrTemplateContentRequired if both BodyTemplate and TextTemplate are empty.
func (t *ExternalTemplate) Validate() error {
	if t.BodyTemplate == "" && t.TextTemplate == "" {
		return ErrTemplateContentRequired
	}

	return nil
}

// RenderedEmail is the result of rendering an external template with branding.
// It contains the fully rendered subject, HTML body, and plain text body
// ready for sending via SendRendered.
type RenderedEmail struct {
	// Subject is the rendered email subject line.
	Subject string

	// HTML is the complete HTML email document including the branded shell.
	HTML string

	// Text is the rendered plain text version of the email.
	Text string
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -run "TestExternalTemplate_Validate" -v ./...`
Expected: PASS — all 4 subtests green

- [ ] **Step 5: Commit**

```bash
git add external_template.go external_template_test.go
git commit -m "feat(external): add ExternalTemplate and RenderedEmail types with validation"
```

---

### Task 3: Data Merging

**Files:**
- Create: `render_external.go`
- Create: `render_external_test.go`

- [ ] **Step 1: Write the failing tests for data merging**

Create `render_external_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -run "TestMergeData" -v ./...`
Expected: FAIL — `mergeData` not defined

- [ ] **Step 3: Implement mergeData**

Create `render_external.go`:

```go
// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

// mergeData merges template defaults with call-site data.
// Call-site data takes precedence over defaults.
// The Branding key is always set from the resolved branding and cannot be overridden.
func mergeData(defaults, data map[string]any, branding Branding) map[string]any {
	merged := make(map[string]any)

	for k, v := range defaults {
		merged[k] = v
	}

	for k, v := range data {
		merged[k] = v
	}

	// Branding is always injected and cannot be overridden
	merged["Branding"] = branding

	return merged
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -run "TestMergeData" -v ./...`
Expected: PASS — all 5 subtests green

- [ ] **Step 5: Commit**

```bash
git add render_external.go render_external_test.go
git commit -m "feat(external): add data merging for external template rendering"
```

---

### Task 4: Template String Rendering Helpers

**Files:**
- Modify: `render_external.go`
- Modify: `render_external_test.go`

- [ ] **Step 1: Write the failing tests for template rendering helpers**

Append to `render_external_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -run "TestRender(Text|HTML)Tmpl" -v ./...`
Expected: FAIL — `renderTextTmpl` and `renderHTMLTmpl` not defined

- [ ] **Step 3: Implement rendering helpers**

Add to `render_external.go`:

```go
import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	texttemplate "text/template"

	"github.com/Masterminds/sprig/v3"
)

// renderTextTmpl renders a Go text/template string with data and Sprig functions.
// Returns empty string without error if tmplStr is empty.
func renderTextTmpl(tmplStr string, data any) (string, error) {
	if tmplStr == "" {
		return "", nil
	}

	t, err := texttemplate.New("").Option("missingkey=error").Funcs(sprig.TxtFuncMap()).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse text template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute text template: %w", err)
	}

	return buf.String(), nil
}

// renderHTMLTmpl renders a Go html/template string with data and Sprig functions.
// HTML-escapes values by default for XSS protection.
// Returns empty string without error if tmplStr is empty.
func renderHTMLTmpl(tmplStr string, data any) (string, error) {
	if tmplStr == "" {
		return "", nil
	}

	t, err := htmltemplate.New("").Option("missingkey=error").Funcs(sprig.HtmlFuncMap()).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse html template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute html template: %w", err)
	}

	return buf.String(), nil
}
```

Note: Make sure the import block at the top of `render_external.go` includes all necessary imports. The full file should have one `import` block combining the imports from Step 3 of Task 3 with these new ones.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -run "TestRender(Text|HTML)Tmpl" -v ./...`
Expected: PASS — all 9 subtests green

- [ ] **Step 5: Commit**

```bash
git add render_external.go render_external_test.go
git commit -m "feat(external): add text and html template rendering helpers with Sprig"
```

---

### Task 5: HTML Email Shell Template

**Files:**
- Create: `shell.go`
- Modify: `render_external.go`
- Modify: `render_external_test.go`

- [ ] **Step 1: Write the failing tests for shell rendering**

Append to `render_external_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -run "TestRenderShell" -v ./...`
Expected: FAIL — `renderShell` not defined

- [ ] **Step 3: Create the HTML email shell template**

Create `shell.go`:

```go
// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

// emailShellTemplate is the HTML email shell template that wraps body content
// with a branded header, footer, and styling. Uses table layout and inline
// styles for maximum email client compatibility (Outlook, Gmail, Yahoo, etc.).
//
// Template data:
//   - .Body: rendered body HTML content (pre-rendered, inserted as-is)
//   - .Preheader: preview text shown in email client list view
//   - .Branding: resolved Branding struct with all color/font/company values
const emailShellTemplate = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
</head>
<body style="margin: 0; padding: 0; background-color: {{.Branding.BackgroundColor}}; font-family: {{.Branding.FontFamily}}; color: {{.Branding.TextColor}}; -webkit-text-size-adjust: 100%; -ms-text-size-adjust: 100%;">
{{- if .Preheader}}
<div style="display: none; max-height: 0; overflow: hidden;">{{.Preheader}}</div>
<div style="display: none; max-height: 0px; overflow: hidden;">&#847;&zwnj;&nbsp;&#847;&zwnj;&nbsp;&#847;&zwnj;&nbsp;&#847;&zwnj;&nbsp;&#847;&zwnj;&nbsp;&#847;&zwnj;&nbsp;</div>
{{- end}}
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="background-color: {{.Branding.BackgroundColor}};">
<tr>
<td align="center" style="padding: 20px 0;">
<table role="presentation" width="600" cellpadding="0" cellspacing="0" border="0" style="max-width: 600px; width: 100%;">
<!-- Header -->
<tr>
<td align="center" style="padding: 20px 40px; border-bottom: 3px solid {{.Branding.PrimaryColor}};">
{{- if .Branding.LogoURL}}
<img src="{{.Branding.LogoURL}}" alt="{{.Branding.BrandName}}" style="max-height: 40px; width: auto;" />
{{- else}}
<span style="font-size: 24px; font-weight: bold; color: {{.Branding.PrimaryColor}}; font-family: {{.Branding.FontFamily}};">{{.Branding.BrandName}}</span>
{{- end}}
</td>
</tr>
<!-- Content -->
<tr>
<td style="background-color: #ffffff; padding: 40px 48px;">
{{.Body}}
</td>
</tr>
<!-- Footer -->
<tr>
<td style="padding: 20px 40px;">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
<tr>
<td style="border-top: 1px solid #e2e8f0; padding-top: 20px; font-size: 12px; color: #64748b; line-height: 1.5;">
<p style="margin: 0 0 8px 0;">If you have any questions, contact us at <a href="mailto:{{.Branding.SupportEmail}}" style="color: {{.Branding.LinkColor}};">{{.Branding.SupportEmail}}</a>.</p>
<p style="margin: 0 0 8px 0;">{{.Branding.CompanyName}}</p>
{{- if .Branding.CompanyAddress}}
<p style="margin: 0; white-space: pre-line;">{{.Branding.CompanyAddress}}</p>
{{- end}}
</td>
</tr>
</table>
</td>
</tr>
</table>
</td>
</tr>
</table>
</body>
</html>`
```

- [ ] **Step 4: Implement renderShell function**

Add to `render_external.go`:

```go
// shellData holds the data passed to the email shell template.
type shellData struct {
	Body      htmltemplate.HTML
	Preheader string
	Branding  Branding
}

// renderShell wraps rendered body HTML in the branded email shell.
// The body parameter is treated as pre-rendered trusted HTML.
func renderShell(body, preheader string, branding Branding) (string, error) {
	t, err := htmltemplate.New("shell").Parse(emailShellTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse shell template: %w", err)
	}

	data := shellData{
		Body:      htmltemplate.HTML(body), // #nosec G203 — body is pre-rendered by renderHTMLTmpl
		Preheader: preheader,
		Branding:  branding,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute shell template: %w", err)
	}

	return buf.String(), nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test -run "TestRenderShell" -v ./...`
Expected: PASS — all 11 subtests green

- [ ] **Step 6: Commit**

```bash
git add shell.go render_external.go render_external_test.go
git commit -m "feat(external): add HTML email shell template with branding support"
```

---

### Task 6: Full RenderTemplate Pipeline

**Files:**
- Modify: `render_external.go`
- Modify: `render_external_test.go`

- [ ] **Step 1: Write the failing tests for RenderTemplate**

Append to `render_external_test.go`:

```go
func TestRenderTemplate(t *testing.T) {
	t.Run("full pipeline with body and text", func(t *testing.T) {
		tmpl := ExternalTemplate{
			SubjectTemplate:   "Welcome to {{.OrgName}}",
			PreheaderTemplate: "You have been invited to {{.OrgName}}",
			BodyTemplate:      `<h1>Hello {{.Name}}</h1><p>Welcome to {{.OrgName}}!</p>`,
			TextTemplate:      "Hello {{.Name}}, welcome to {{.OrgName}}!",
		}
		data := map[string]any{
			"Name":    "Max",
			"OrgName": "Acme Corp",
		}

		rendered, err := RenderTemplate(tmpl, nil, data)
		require.NoError(t, err)
		require.Equal(t, "Welcome to Acme Corp", rendered.Subject)
		require.Contains(t, rendered.HTML, "Hello Max")
		require.Contains(t, rendered.HTML, "Welcome to Acme Corp!")
		require.Contains(t, rendered.HTML, "<!DOCTYPE html")
		require.Contains(t, rendered.HTML, DefaultBrandName)
		require.Contains(t, rendered.Text, "Hello Max, welcome to Acme Corp!")
	})

	t.Run("branding accessible in body template", func(t *testing.T) {
		tmpl := ExternalTemplate{
			BodyTemplate: `<a style="background-color: {{.Branding.ButtonColor}}; color: {{.Branding.ButtonTextColor}};">Click</a>`,
		}

		rendered, err := RenderTemplate(tmpl, nil, nil)
		require.NoError(t, err)
		require.Contains(t, rendered.HTML, DefaultButtonColor)
		require.Contains(t, rendered.HTML, DefaultButtonTextColor)
	})

	t.Run("custom branding", func(t *testing.T) {
		tmpl := ExternalTemplate{
			BodyTemplate: `<p>Hello</p>`,
		}
		branding := &Branding{
			BrandName:    "Acme",
			PrimaryColor: "#ff0000",
		}

		rendered, err := RenderTemplate(tmpl, branding, nil)
		require.NoError(t, err)
		require.Contains(t, rendered.HTML, "Acme")
		require.Contains(t, rendered.HTML, "#ff0000")
	})

	t.Run("defaults merged with data", func(t *testing.T) {
		tmpl := ExternalTemplate{
			BodyTemplate: `<p>{{.Greeting}} {{.Name}}</p>`,
			Defaults:     map[string]any{"Greeting": "Hello"},
		}
		data := map[string]any{"Name": "Max"}

		rendered, err := RenderTemplate(tmpl, nil, data)
		require.NoError(t, err)
		require.Contains(t, rendered.HTML, "Hello Max")
	})

	t.Run("body only - no text", func(t *testing.T) {
		tmpl := ExternalTemplate{
			SubjectTemplate: "Test",
			BodyTemplate:    "<p>Body only</p>",
		}

		rendered, err := RenderTemplate(tmpl, nil, nil)
		require.NoError(t, err)
		require.Contains(t, rendered.HTML, "Body only")
		require.Empty(t, rendered.Text)
	})

	t.Run("text only - no body", func(t *testing.T) {
		tmpl := ExternalTemplate{
			SubjectTemplate: "Test",
			TextTemplate:    "Text only",
		}

		rendered, err := RenderTemplate(tmpl, nil, nil)
		require.NoError(t, err)
		require.Empty(t, rendered.HTML)
		require.Equal(t, "Text only", rendered.Text)
	})

	t.Run("empty subject returns empty string", func(t *testing.T) {
		tmpl := ExternalTemplate{
			BodyTemplate: "<p>Content</p>",
		}

		rendered, err := RenderTemplate(tmpl, nil, nil)
		require.NoError(t, err)
		require.Empty(t, rendered.Subject)
	})

	t.Run("invalid body template returns error when text also fails", func(t *testing.T) {
		tmpl := ExternalTemplate{
			BodyTemplate: "{{.Missing}",
			TextTemplate: "{{.Missing}",
		}

		_, err := RenderTemplate(tmpl, nil, nil)
		require.Error(t, err)
	})

	t.Run("invalid body but valid text succeeds", func(t *testing.T) {
		tmpl := ExternalTemplate{
			BodyTemplate: "{{.Missing}",
			TextTemplate: "Fallback text",
		}

		rendered, err := RenderTemplate(tmpl, nil, nil)
		require.NoError(t, err)
		require.Empty(t, rendered.HTML)
		require.Equal(t, "Fallback text", rendered.Text)
	})

	t.Run("validation error for empty templates", func(t *testing.T) {
		tmpl := ExternalTemplate{}

		_, err := RenderTemplate(tmpl, nil, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrTemplateContentRequired)
	})
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -run "TestRenderTemplate" -v ./...`
Expected: FAIL — `RenderTemplate` not defined

- [ ] **Step 3: Implement RenderTemplate**

Add to `render_external.go`:

```go
// RenderTemplate renders an external template with branding and data into a ready-to-send email.
//
// The rendering pipeline:
//  1. Validate the template (at least one content template required)
//  2. Merge tmpl.Defaults with data (data takes precedence)
//  3. Resolve branding (fill zero values with Kopexa defaults)
//  4. Render SubjectTemplate and PreheaderTemplate with text/template
//  5. Render BodyTemplate with html/template (XSS protection)
//  6. Render TextTemplate with text/template
//  7. Wrap rendered body in the branded email shell
//
// If branding is nil, Kopexa defaults are used.
// If both BodyTemplate and TextTemplate fail to render, an error is returned.
// If only one fails, the successful result is returned without error.
//
// Example:
//
//	tmpl := comms.ExternalTemplate{
//	    SubjectTemplate: "Welcome to {{.OrgName}}",
//	    BodyTemplate:    "<h1>Hello {{.Name}}</h1>",
//	    TextTemplate:    "Hello {{.Name}}",
//	}
//	rendered, err := comms.RenderTemplate(tmpl, nil, map[string]any{
//	    "Name":    "Max",
//	    "OrgName": "Acme Corp",
//	})
func RenderTemplate(tmpl ExternalTemplate, branding *Branding, data map[string]any) (*RenderedEmail, error) {
	if err := tmpl.Validate(); err != nil {
		return nil, err
	}

	resolved := resolveBranding(branding)
	merged := mergeData(tmpl.Defaults, data, resolved)

	// Render subject (text/template, no HTML escaping)
	subject, err := renderTextTmpl(tmpl.SubjectTemplate, merged)
	if err != nil {
		return nil, fmt.Errorf("failed to render subject: %w", err)
	}

	// Render preheader (text/template)
	preheader, err := renderTextTmpl(tmpl.PreheaderTemplate, merged)
	if err != nil {
		return nil, fmt.Errorf("failed to render preheader: %w", err)
	}

	// Render body (html/template for XSS protection)
	bodyHTML, bodyErr := renderHTMLTmpl(tmpl.BodyTemplate, merged)

	// Render plain text (text/template)
	textContent, textErr := renderTextTmpl(tmpl.TextTemplate, merged)

	// Both failed → error
	if bodyErr != nil && textErr != nil {
		return nil, fmt.Errorf(
			"%w: body error: %v, text error: %v",
			ErrBothTemplatesFailed, bodyErr, textErr,
		)
	}

	// Wrap body in shell (only if body rendered successfully)
	var fullHTML string
	if bodyErr == nil && bodyHTML != "" {
		fullHTML, err = renderShell(bodyHTML, preheader, resolved)
		if err != nil {
			return nil, fmt.Errorf("failed to render shell: %w", err)
		}
	}

	return &RenderedEmail{
		Subject: subject,
		HTML:    fullHTML,
		Text:    textContent,
	}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -run "TestRenderTemplate" -v ./...`
Expected: PASS — all 10 subtests green

- [ ] **Step 5: Run all tests to check nothing is broken**

Run: `go test -v ./...`
Expected: PASS — all existing and new tests pass

- [ ] **Step 6: Commit**

```bash
git add render_external.go render_external_test.go
git commit -m "feat(external): add RenderTemplate with full rendering pipeline"
```

---

### Task 7: SendRendered & SendFromTemplate

**Files:**
- Create: `send_external.go`
- Create: `send_external_test.go`

- [ ] **Step 1: Write the failing tests for SendRendered**

Create `send_external_test.go`:

```go
// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"errors"
	"testing"

	"github.com/kopexa-grc/comms/v2/driver"
	"github.com/kopexa-grc/comms/v2/driver/mock"
	"github.com/stretchr/testify/require"
)

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
			return errors.New("driver error")
		}
		c := New(WithDriver(d), WithFrom("noreply@example.com"))
		recipient := Recipient{Email: "user@example.com"}

		rendered := &RenderedEmail{
			Subject: "Subject",
			HTML:    "<p>body</p>",
		}

		err := c.SendRendered(context.Background(), recipient, rendered)
		require.Error(t, err)
		require.Contains(t, err.Error(), "driver error")
	})
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -run "TestComms_SendRendered" -v ./...`
Expected: FAIL — `SendRendered` not defined

- [ ] **Step 3: Implement SendRendered**

Create `send_external.go`:

```go
// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

import (
	"context"
	"errors"
	"fmt"

	"github.com/kopexa-grc/comms/v2/driver"
)

var (
	ErrRenderedRequired = errors.New("rendered email is required")
	ErrSubjectRequired  = errors.New("rendered email subject is required")
	ErrBodyRequired     = errors.New("rendered email must have HTML or Text body")
)

// SendRendered sends a pre-rendered email to the given recipient.
//
// Use this after RenderTemplate when you need to separate rendering from sending,
// e.g., for previews, caching, or bulk sends.
//
// Validates before sending:
//   - rendered must not be nil
//   - rendered.Subject must not be empty
//   - rendered.HTML or rendered.Text must be non-empty
//   - Recipient email must be valid
func (c *Comms) SendRendered(ctx context.Context, r Recipient, rendered *RenderedEmail) error {
	if rendered == nil {
		return fmt.Errorf("%w", ErrRenderedRequired)
	}

	if rendered.Subject == "" {
		return fmt.Errorf("%w", ErrSubjectRequired)
	}

	if rendered.HTML == "" && rendered.Text == "" {
		return fmt.Errorf("%w", ErrBodyRequired)
	}

	if err := r.Validate(); err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}

	msg := driver.Message{
		From:    c.config.From,
		To:      []string{r.String()},
		Subject: rendered.Subject,
		HTML:    rendered.HTML,
		Text:    rendered.Text,
	}

	if err := c.config.Driver.Send(ctx, msg); err != nil {
		return fmt.Errorf("failed to send rendered email: %w", err)
	}

	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -run "TestComms_SendRendered" -v ./...`
Expected: PASS — all 6 subtests green

- [ ] **Step 5: Write the failing tests for SendFromTemplate**

Append to `send_external_test.go`:

```go
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
```

- [ ] **Step 6: Run tests to verify they fail**

Run: `go test -run "TestComms_SendFromTemplate" -v ./...`
Expected: FAIL — `SendFromTemplate` not defined

- [ ] **Step 7: Implement SendFromTemplate**

Add to `send_external.go`:

```go
// SendFromTemplate is a convenience method that renders an external template and sends
// it in one call. Equivalent to calling RenderTemplate followed by SendRendered.
//
// Example:
//
//	tmpl := comms.ExternalTemplate{
//	    SubjectTemplate: "Welcome to {{.OrgName}}",
//	    BodyTemplate:    "<h1>Hello {{.Name}}</h1>",
//	    TextTemplate:    "Hello {{.Name}}",
//	}
//	err := c.SendFromTemplate(ctx, recipient, tmpl, nil, map[string]any{
//	    "Name":    "Max",
//	    "OrgName": "Acme Corp",
//	})
func (c *Comms) SendFromTemplate(ctx context.Context, r Recipient, tmpl ExternalTemplate, branding *Branding, data map[string]any) error {
	rendered, err := RenderTemplate(tmpl, branding, data)
	if err != nil {
		return err
	}

	return c.SendRendered(ctx, r, rendered)
}
```

- [ ] **Step 8: Run tests to verify they pass**

Run: `go test -run "TestComms_SendFromTemplate" -v ./...`
Expected: PASS — all 3 subtests green

- [ ] **Step 9: Run full test suite**

Run: `go test -v ./...`
Expected: PASS — all tests pass (existing + new)

- [ ] **Step 10: Commit**

```bash
git add send_external.go send_external_test.go
git commit -m "feat(external): add SendRendered and SendFromTemplate methods"
```

---

### Task 8: GoDoc Examples

**Files:**
- Create: `example_external_test.go`

- [ ] **Step 1: Create testable GoDoc examples**

Create `example_external_test.go`:

```go
// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms_test

import (
	"context"
	"fmt"

	"github.com/kopexa-grc/comms/v2"
	"github.com/kopexa-grc/comms/v2/driver/mock"
)

func ExampleRenderTemplate() {
	tmpl := comms.ExternalTemplate{
		SubjectTemplate: "Welcome to {{.OrgName}}",
		BodyTemplate:    "<h1>Hello {{.Name}}</h1><p>Welcome aboard!</p>",
		TextTemplate:    "Hello {{.Name}}, welcome aboard!",
	}

	rendered, err := comms.RenderTemplate(tmpl, nil, map[string]any{
		"Name":    "Max",
		"OrgName": "Acme Corp",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(rendered.Subject)
	fmt.Println(rendered.Text)
	// Output:
	// Welcome to Acme Corp
	// Hello Max, welcome aboard!
}

func ExampleRenderTemplate_withBranding() {
	tmpl := comms.ExternalTemplate{
		SubjectTemplate: "Hello",
		BodyTemplate:    `<a style="background-color: {{.Branding.ButtonColor}};">Click</a>`,
	}

	branding := &comms.Branding{
		BrandName:    "Acme Corp",
		PrimaryColor: "#ff6600",
		ButtonColor:  "#cc0000",
	}

	rendered, err := comms.RenderTemplate(tmpl, branding, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(rendered.Subject)
	// Output:
	// Hello
}

func ExampleComms_SendFromTemplate() {
	d := mock.NewDriver()
	c := comms.New(
		comms.WithDriver(d),
		comms.WithFrom("noreply@example.com"),
	)

	recipient := comms.NewRecipient("user@example.com", "Max", "Mustermann")

	tmpl := comms.ExternalTemplate{
		SubjectTemplate: "Welcome to {{.OrgName}}",
		BodyTemplate:    "<h1>Hello {{.Name}}</h1>",
		TextTemplate:    "Hello {{.Name}}",
	}

	err := c.SendFromTemplate(context.Background(), recipient, tmpl, nil, map[string]any{
		"Name":    "Max",
		"OrgName": "Acme Corp",
	})
	if err != nil {
		panic(err)
	}

	msg, _ := d.LastMessage()
	fmt.Println(msg.Subject)
	// Output:
	// Welcome to Acme Corp
}
```

- [ ] **Step 2: Run examples to verify they pass**

Run: `go test -run "Example" -v ./...`
Expected: PASS — all 3 examples pass

- [ ] **Step 3: Commit**

```bash
git add example_external_test.go
git commit -m "docs: add testable GoDoc examples for external templates"
```

---

### Task 9: README Documentation

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Add External Templates section to README**

Insert after the existing "Templates" section (after line 99) in `README.md`:

```markdown
### External Templates

External templates allow you to render and send emails from template strings loaded at runtime (e.g., from a database). They are wrapped in a branded email shell automatically.

#### Quick Start

```go
// Define template (typically loaded from database)
tmpl := comms.ExternalTemplate{
    SubjectTemplate: "Welcome to {{.OrgName}}",
    BodyTemplate:    "<h1>Hello {{.Name}}</h1><p>Welcome aboard!</p>",
    TextTemplate:    "Hello {{.Name}}, welcome aboard!",
}

// Optional: customize branding (nil = Kopexa defaults)
branding := &comms.Branding{
    BrandName:    "Acme Corp",
    LogoURL:      "https://acme.com/logo.png",
    PrimaryColor: "#ff6600",
    ButtonColor:  "#cc0000",
    SupportEmail: "help@acme.com",
}

// Render only (for previews)
rendered, err := comms.RenderTemplate(tmpl, branding, map[string]any{
    "Name":    "Max",
    "OrgName": "Acme Corp",
})

// Or render + send in one call
err = c.SendFromTemplate(ctx, recipient, tmpl, branding, data)
```

#### Template Syntax

Templates use Go's [template syntax](https://pkg.go.dev/text/template) with [Sprig v3](https://masterminds.github.io/sprig/) functions available.

**Body templates** (`BodyTemplate`) are rendered with `html/template` for XSS protection. **Subject** and **Text** templates use `text/template` (no HTML escaping).

#### Using Branding in Templates

Branding values are available in all templates via the `.Branding` key:

```html
<a href="{{.URL}}" style="background-color: {{.Branding.ButtonColor}}; color: {{.Branding.ButtonTextColor}}; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">
    Click here
</a>
```

#### Template Defaults

Templates support a `Defaults` map for base-layer variables. Call-site data takes precedence:

```go
tmpl := comms.ExternalTemplate{
    BodyTemplate: "<p>Visit {{.AppURL}}</p>",
    Defaults:     map[string]any{"AppURL": "https://app.example.com"},
}

// Call-site data overrides defaults
data := map[string]any{"AppURL": "https://custom.example.com"}
```

#### Branding Defaults

| Field | Default |
|-------|---------|
| BrandName | Kopexa |
| PrimaryColor | #2563eb |
| BackgroundColor | #f1f5f9 |
| TextColor | #0f172a |
| ButtonColor | #2563eb |
| ButtonTextColor | #ffffff |
| LinkColor | #2563eb |
| FontFamily | Helvetica, Arial, sans-serif |
| CompanyName | Kopexa GmbH |
| SupportEmail | support@kopexa.com |

#### Two-Phase API

For advanced use cases (previews, caching, bulk sends), use the two-phase API:

```go
// Phase 1: Render (pure, no driver needed)
rendered, err := comms.RenderTemplate(tmpl, branding, data)

// Phase 2: Send (uses configured driver)
err = c.SendRendered(ctx, recipient, rendered)
```
```

- [ ] **Step 2: Verify README renders correctly**

Run: `head -200 README.md` — visually confirm the structure looks correct.

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs: add External Templates section to README"
```

---

### Task 10: License Headers & Lint Pass

**Files:**
- All new `.go` files

- [ ] **Step 1: Verify license headers**

Run: `make license/headers/check`
Expected: PASS — all files have correct headers.
If any are missing, run: `make license/headers/apply`

- [ ] **Step 2: Run linter**

Run: `make lint`
Expected: PASS — no lint errors.
If there are issues, fix them in the relevant files.

- [ ] **Step 3: Run full test suite**

Run: `make test/unit`
Expected: PASS — all tests pass.

- [ ] **Step 4: Commit any fixes**

If license headers or lint fixes were needed:

```bash
git add -A
git commit -m "chore: fix license headers and lint issues"
```

---
