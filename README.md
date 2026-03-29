# Comms

[![Go Report Card](https://goreportcard.com/badge/github.com/kopexa-grc/comms)](https://goreportcard.com/report/github.com/kopexa-grc/comms)
[![License](https://img.shields.io/badge/License-BUSL--1.1-blue.svg)](LICENSE)

> **Note**: This is an internal library used within the Kopexa core ecosystem. It is not intended for public use.

Comms is a flexible and extensible email communication library for Go that enables easy integration with various email services. It is primarily used within Kopexa's internal services and applications.

## Features

- 🚀 Easy integration with various email services
- 📧 Support for HTML and text emails
- 📎 Attachment support
- 🏷️ Tagging system for email categorization
- 🔒 Email address validation
- 🧪 Comprehensive test coverage

## Installation

```bash
go get github.com/kopexa-grc/comms
```

> **Warning**: This library is licensed under BUSL-1.1 and is intended for internal use within Kopexa. Please ensure you have the appropriate permissions and licensing before using this library.

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/kopexa-grc/comms"
    "github.com/kopexa-grc/comms/driver/resend"
)

func main() {
    // Configure Resend driver
    driver := resend.New("your-api-key")

    // Create Comms instance
    c := comms.New(
        comms.WithDriver(driver),
        comms.WithFrom("noreply@example.com"),
    )

    // Create recipient
    recipient := comms.NewRecipient(
        "user@example.com",
        "John",
        "Doe",
    )

    // Send email
    err := c.SendVerifyEmail(context.Background(), recipient, "123456")
    if err != nil {
        log.Fatal(err)
    }
}
```

## Drivers

### Resend

The Resend driver enables sending emails through the Resend service.

```go
driver := resend.New("your-api-key")
```

### Mock

The Mock driver is designed for testing purposes and simulates email sending.

```go
driver := mock.NewDriver()
```

## Templates

Comms supports HTML and text templates for emails. Templates are written in Go template syntax and can contain dynamic data.

### Template Structure

```
templates/
├── verify_email.html
└── verify_email.txt
```

### Template Usage

```go
text, html, err := comms.Render("verify_email", data)
```

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

## Configuration

### Options

- `WithDriver(driver.Driver)`: Sets the email driver
- `WithFrom(string)`: Sets the sender email address

### Example

```go
c := comms.New(
    comms.WithDriver(driver),
    comms.WithFrom("noreply@example.com"),
)
```

## Development

### Prerequisites

- Go 1.21 or higher
- Make

### Build

```bash
make build
```

### Tests

```bash
make test
```

### Linting

```bash
make lint
```

## License

This project is licensed under the [BUSL-1.1](LICENSE) License. This license is specifically designed for internal use within Kopexa and its ecosystem.

## Security

For security issues, please read our [SECURITY.md](SECURITY.md) file.

## Contributing

Contributions are welcome! Please read our [Contributing Guidelines](CONTRIBUTING.md) for details.

## Support

For questions or issues, please create an issue in the GitHub repository.
