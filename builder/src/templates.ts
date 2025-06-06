/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import ForgotPassword from "./auth/forgot-password";
import { VerifyEmail } from "./auth/verify-email";
import Welcome from "./auth/welcome";
import OrgCreated from "./org/created";
import OrgInvite from "./org/invite";

export const templates = [
  {
    component: VerifyEmail,
    name: "verify-email",
  },
  {
    component: Welcome,
    name: "welcome",
  },
  {
    component: ForgotPassword,
    name: "forgot-password",
  },
  {
    component: OrgCreated,
    name: "org-created",
  },
  {
    component: OrgInvite,
    name: "org-invite",
  },
];
