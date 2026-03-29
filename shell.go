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
