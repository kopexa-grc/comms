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
