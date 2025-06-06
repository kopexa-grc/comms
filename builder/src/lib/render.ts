/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

export function renderOrFallback(value: string | undefined | null, fallback: string): string {
  if (!value || value.trim() === "") {
    return fallback;
  }
  return value;
}