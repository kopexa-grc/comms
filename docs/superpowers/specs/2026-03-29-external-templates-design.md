# External Templates Design Spec

**Date:** 2026-03-29
**Status:** Approved

## Summary

Extend the `comms` library with the ability to render and send emails from externally-provided templates (e.g., stored in a database) with configurable branding. Existing hardcoded templates remain unchanged — this is a purely additive extension.

## Motivation

Kopexa stores email templates in a database (`EmailTemplate` entity) so users can customize them. The `comms` library currently only supports compile-time embedded React Email templates. A new rendering path is needed that accepts template strings and branding configuration at runtime.

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Template source | Caller passes template strings (DB-agnostic) | Comms stays independent of DB schema; caller controls loading, caching, fallback |
| Rendering | Comms renders Go template strings + applies branding shell | Comms already has rendering infrastructure; keeps email HTML expertise in the library |
| HTML shell | Go `html/template` with table layout + inline styles | No build pipeline dependency; full runtime control; max email client compatibility |
| Existing templates | Unchanged, coexist alongside external templates | Additive extension, no breaking changes |
| API style | Two-phase: RenderTemplate (pure) + SendRendered (with driver) | TDD-friendly, enables previews and caching, render and send independently testable |

## New Types

### ExternalTemplate

Holds template strings loaded from an external source (e.g., database).

```go
type ExternalTemplate struct {
    // SubjectTemplate is a Go template string for the email subject line.
    // Example: "{{.ActorName}} invited you to {{.OrgName}}"
    SubjectTemplate string

    // PreheaderTemplate is a Go template string for the email preview text.
    PreheaderTemplate string

    // BodyTemplate is a Go html/template string for the email body content.
    // Rendered INSIDE the email shell — do NOT include DOCTYPE/html/body tags.
    BodyTemplate string

    // TextTemplate is a Go text/template string for the plain text fallback.
    // If empty, no plain text version is included.
    TextTemplate string

    // Defaults are base-layer variables merged with call-site data.
    // Call-site data takes precedence over defaults.
    Defaults map[string]any
}
```

### Branding

Configures the visual appearance of the email shell. Zero values fall back to Kopexa defaults.

```go
type Branding struct {
    BrandName       string // Displayed in header and footer. Default: "Kopexa"
    LogoURL         string // URL to brand logo. If empty, BrandName shown as text.
    PrimaryColor    string // Primary accent color (hex). Default: "#2563eb"
    BackgroundColor string // Email background (hex). Default: "#f1f5f9"
    TextColor       string // Body text color (hex). Default: "#0f172a"
    ButtonColor     string // CTA button background (hex). Default: "#2563eb"
    ButtonTextColor string // CTA button text color (hex). Default: "#ffffff"
    LinkColor       string // Link color (hex). Default: "#2563eb"
    FontFamily      string // CSS font-family. Default: "Helvetica, Arial, sans-serif"

    // Footer content
    CompanyName    string // Legal entity name. Default: "Kopexa GmbH"
    CompanyAddress string // Multiline address for footer.
    SupportEmail   string // Support contact. Default: "support@kopexa.com"
}
```

### RenderedEmail

Result of rendering an external template — ready for sending.

```go
type RenderedEmail struct {
    Subject string
    HTML    string
    Text    string
}
```

## API Methods

### RenderTemplate (package-level function)

```go
func RenderTemplate(tmpl ExternalTemplate, branding *Branding, data map[string]any) (*RenderedEmail, error)
```

Pure function. No driver needed. Renders template strings with data and wraps the body in the branded email shell.

- `branding` is a pointer: `nil` = Kopexa defaults
- `data` is `map[string]any` because external templates have dynamic variables not known at compile time

### SendRendered (method on Comms)

```go
func (c *Comms) SendRendered(ctx context.Context, r Recipient, rendered *RenderedEmail) error
```

Sends a pre-rendered email via the configured driver. Use after `RenderTemplate` when render and send need to be separated (previews, caching, bulk sends).

Validates before sending:
- `rendered` must not be nil
- `rendered.Subject` must not be empty
- `rendered.HTML` or `rendered.Text` must be non-empty (at least one body)
- Recipient is validated via existing `Recipient.Validate()`

### SendFromTemplate (convenience method on Comms)

```go
func (c *Comms) SendFromTemplate(ctx context.Context, r Recipient, tmpl ExternalTemplate, branding *Branding, data map[string]any) error
```

Equivalent to `RenderTemplate` + `SendRendered` in one call.

## Rendering Pipeline

```
1. Merge tmpl.Defaults with data (data wins)
2. Resolve branding (fill zero values with Kopexa defaults)
3. Build template context: merged data + Branding under .Branding key
4. Render SubjectTemplate   (text/template — no HTML escaping)
5. Render PreheaderTemplate (text/template)
6. Render BodyTemplate      (html/template — XSS protection)
7. Render TextTemplate      (text/template)
8. Render Shell template    (html/template) with:
   - .Body = rendered body HTML
   - .Preheader = rendered preheader
   - .Branding = resolved branding
9. Return RenderedEmail{Subject, HTML, Text}
```

**Error handling:**
- If both BodyTemplate and TextTemplate fail to render → return error (analogous to `ErrBothTemplatesFailed`). If only one fails → return the successful one without error.
- If SubjectTemplate is empty → rendered Subject is empty string (no error). This allows preview/rendering without a subject. `SendRendered` validates that Subject is non-empty before sending.
- If BodyTemplate and TextTemplate are both empty strings → `ExternalTemplate.Validate()` returns an error. At least one content template is required.

## HTML Email Shell

The shell is an embedded Go `html/template` string that produces the full HTML email document.

### Structure

```
┌─────────────────────────────────────────┐
│ DOCTYPE + HTML + Head                   │
│  ├─ meta charset, viewport              │
│  └─ style block (from Branding)         │
├─────────────────────────────────────────┤
│ Body (background: Branding.Background)  │
│  ┌───────────────────────────────────┐  │
│  │ Container (max-width: 600px)      │  │
│  │  ┌─────────────────────────────┐  │  │
│  │  │ Header                      │  │  │
│  │  │  Logo (img) or BrandName    │  │  │
│  │  └─────────────────────────────┘  │  │
│  │  ┌─────────────────────────────┐  │  │
│  │  │ Preheader (hidden preview)  │  │  │
│  │  └─────────────────────────────┘  │  │
│  │  ┌─────────────────────────────┐  │  │
│  │  │ Content Card (white bg)     │  │  │
│  │  │  {{ .Body }}                │  │  │
│  │  └─────────────────────────────┘  │  │
│  │  ┌─────────────────────────────┐  │  │
│  │  │ Footer                      │  │  │
│  │  │  CompanyName + Address      │  │  │
│  │  │  SupportEmail               │  │  │
│  │  └─────────────────────────────┘  │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

### Branding application

- **Shell:** BackgroundColor, TextColor, FontFamily, PrimaryColor (header accent), LinkColor (footer links), LogoURL/BrandName, CompanyName, CompanyAddress, SupportEmail
- **Body templates:** Authors access branding via `{{.Branding.ButtonColor}}`, `{{.Branding.ButtonTextColor}}`, etc. for styling CTAs and links

### Implementation

- Table layout for email client compatibility (Outlook, Gmail)
- Inline styles only — no CSS classes, no Tailwind, no `<style>` in head
- Logo: `<img>` if LogoURL set, otherwise BrandName as styled text

## Data Merging

```
tmpl.Defaults = {"SupportEmail": "help@acme.com", "AppURL": "https://app.acme.com"}
data           = {"Name": "Max", "AppURL": "https://custom.acme.com"}
                                    │
                                    ▼
merged = {
    "SupportEmail": "help@acme.com",   ← from Defaults (not in data)
    "AppURL": "https://custom.acme.com", ← from data (wins over Defaults)
    "Name": "Max",                       ← from data
    "Branding": { resolved Branding },   ← injected by RenderTemplate
}
```

Single-level merge. Call-site `data` takes precedence over `tmpl.Defaults`. The `Branding` key is injected automatically and cannot be overridden by data.

## File Structure

All new files — no changes to existing files.

```
comms/
├── branding.go               # Branding type + resolveBranding() + default constants
├── branding_test.go          # Default resolution: nil, partial, full
├── external_template.go      # ExternalTemplate + RenderedEmail types + Validate()
├── external_template_test.go # Validation tests
├── render_external.go        # RenderTemplate() + renderShell() + mergeData()
├── render_external_test.go   # Full render pipeline tests
├── send_external.go          # SendRendered() + SendFromTemplate()
├── send_external_test.go     # Send tests with mock driver
├── shell.go                  # HTML email shell template constant
├── example_external_test.go  # Testable GoDoc examples
```

## TDD Plan

Tests are written first, then implementation, phase by phase:

| Phase | Tests | Implementation |
|-------|-------|----------------|
| 1 | `branding_test.go` — `resolveBranding`: nil→all defaults, partial→merged, full→unchanged | `branding.go` — type + resolveBranding + constants |
| 2 | `external_template_test.go` — `Validate`: both empty→error, body only→ok, text only→ok, both set→ok | `external_template.go` — types + Validate |
| 3 | `render_external_test.go` — data merging: defaults+data, data wins, Branding injected | `render_external.go` — mergeData |
| 4 | `render_external_test.go` — individual template rendering: subject, body, text; Sprig functions; invalid syntax→error | `render_external.go` — template rendering helpers |
| 5 | `render_external_test.go` — shell rendering: branding colors in HTML, logo vs. text, preheader, footer | `shell.go` + `render_external.go` — renderShell |
| 6 | `render_external_test.go` — full pipeline: `RenderTemplate()` end-to-end, `{{.Branding.ButtonColor}}` in body | `render_external.go` — RenderTemplate |
| 7 | `send_external_test.go` — `SendRendered`: message built correctly, driver called | `send_external.go` — SendRendered |
| 8 | `send_external_test.go` — `SendFromTemplate`: render+send integration, error propagation | `send_external.go` — SendFromTemplate |

## Documentation Plan

| Artifact | Location | Content |
|----------|----------|---------|
| GoDoc comments | All public types, methods, fields | Usage docs with examples |
| Testable examples | `example_external_test.go` | Runnable examples shown in GoDoc |
| README extension | `README.md` — new "External Templates" section | Quickstart, branding defaults table, template syntax guide, available Sprig functions |
| Template authoring guide | README subsection | Available variables, how to use `{{.Branding.*}}` for buttons/links, best practices |

## Branding Defaults Reference

| Field | Default Value |
|-------|--------------|
| BrandName | `"Kopexa"` |
| PrimaryColor | `"#2563eb"` |
| BackgroundColor | `"#f1f5f9"` |
| TextColor | `"#0f172a"` |
| ButtonColor | `"#2563eb"` |
| ButtonTextColor | `"#ffffff"` |
| LinkColor | `"#2563eb"` |
| FontFamily | `"Helvetica, Arial, sans-serif"` |
| CompanyName | `"Kopexa GmbH"` |
| CompanyAddress | `"Schauenburgerstr. 116\n24118 Kiel\nGermany"` |
| SupportEmail | `"support@kopexa.com"` |
