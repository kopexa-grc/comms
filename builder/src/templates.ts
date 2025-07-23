/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { VendorAssessmentRequest } from "./assessment/vendor-request";
import ForgotPassword from "./auth/forgot-password";
import PasswordResetSuccess from "./auth/password-reset-success";
import { VerifyEmail } from "./auth/verify-email";
import Welcome from "./auth/welcome";
import OrgCreated from "./org/created";
import OrgInvite from "./org/invite";
import InviteAccepted from "./org/invite-accepted";
import Subscribe from "./subscribe";

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
  {
    component: PasswordResetSuccess,
    name: "password-reset-success",
  },
  {
    component: InviteAccepted,
    name: "org-invite-accepted",
  },
  {
    component: VendorAssessmentRequest,
    name: "vendor-assessment-request",
  },
  {
    component: Subscribe,
    name: "subscribe",
  },
] as const;
