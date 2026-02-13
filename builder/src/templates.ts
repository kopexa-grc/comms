/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { VendorAssessmentRequest } from "./assessment/vendor-request";
import { VendorSurveyOtp } from "./assessment/vendor-survey-otp";
import ForgotPassword from "./auth/forgot-password";
import PasswordResetSuccess from "./auth/password-reset-success";
import RecoveryCodesRegenerated from "./auth/recovery-codes-regenerated";
import { VerifyEmail } from "./auth/verify-email";
import Welcome from "./auth/welcome";
import OrgCreated from "./org/created";
import OrgInvite from "./org/invite";
import InviteAccepted from "./org/invite-accepted";
import ReviewOverdue from "./review/review-overdue";
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
    component: VendorSurveyOtp,
    name: "vendor-survey-otp",
  },
  {
    component: Subscribe,
    name: "subscribe",
  },
  {
    component: ReviewOverdue,
    name: "review-overdue",
  },
  {
    component: RecoveryCodesRegenerated,
    name: "recovery-codes-regenerated",
  },
] as const;
