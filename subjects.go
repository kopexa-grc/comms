// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package comms

// subjects contains localized email subjects keyed by template name and language.
var subjects = map[string]map[string]string{
	"dsr-receipt-confirmation": {
		"en": "Confirmation of Your Data Subject Request - %s",
		"de": "Bestätigung deiner Betroffenenanfrage - %s",
	},
	"forgot-password": {
		"en": "Reset your password",
		"de": "Setze dein Passwort zurück",
	},
	"incident-deadline-overdue": {
		"en": "URGENT: Incident reporting deadline overdue: %s",
		"de": "DRINGEND: Meldefrist überschritten: %s",
	},
	"incident-deadline-reminder": {
		"en": "Incident reporting deadline approaching: %s",
		"de": "Meldefrist läuft bald ab: %s",
	},
	"org-created": {
		"en": "Your organization has been created",
		"de": "Deine Organisation wurde erstellt",
	},
	"org-invite": {
		"en": "You have been invited to join %s on Kopexa",
		"de": "Du wurdest eingeladen, %s auf Kopexa beizutreten",
	},
	"org-invite-accepted": {
		"en": "%s has accepted your invitation to join %s",
		"de": "%s hat deine Einladung zu %s angenommen",
	},
	"org-transfer-sender-confirm": {
		"en": "Confirm ownership transfer for %s",
		"de": "Bestätige die Übertragung von %s",
	},
	"org-transfer-receiver-invite": {
		"en": "Take over the organization %s",
		"de": "Übernimm die Organisation %s",
	},
	"password-reset-success": {
		"en": "Your password has been successfully reset",
		"de": "Dein Passwort wurde erfolgreich zurückgesetzt",
	},
	"recovery-codes-regenerated": {
		"en": "Your recovery codes have been regenerated",
		"de": "Deine Wiederherstellungscodes wurden neu generiert",
	},
	"review-overdue": {
		"en": "Review overdue notification",
		"de": "Überprüfung überfällig",
	},
	"review-pending": {
		"en": "Review required notification",
		"de": "Überprüfung erforderlich",
	},
	"subscribe": {
		"en": "Thank you for subscribing",
		"de": "Danke für deine Anmeldung",
	},
	"vendor-assessment-request": {
		"en": "%s has requested a vendor assessment from you",
		"de": "%s hat eine Lieferantenbewertung von dir angefordert",
	},
	"vendor-survey-otp": {
		"en": "Your verification code for the vendor assessment",
		"de": "Dein Bestätigungscode für die Lieferantenbewertung",
	},
	"verify-email": {
		"en": "Verify your email address",
		"de": "Bestätige deine E-Mail-Adresse",
	},
	"welcome": {
		"en": "Welcome to Kopexa",
		"de": "Willkommen bei Kopexa",
	},
}

// Subject returns the localized subject for the given template and language.
// Falls back to English if the language is not found.
func Subject(template, lang string) string {
	if s, ok := subjects[template]; ok {
		if subj, ok := s[lang]; ok {
			return subj
		}
		return s["en"]
	}
	return ""
}
