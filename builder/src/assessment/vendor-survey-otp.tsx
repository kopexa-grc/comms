/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Text } from "@react-email/components";
import { Base } from "../partials/base";
import { Container } from "../partials/container";
import { Header } from "../partials/header";
import { Box } from "../partials/box";

type VendorSurveyOtpProps = {
  displayName?: string;
  otpCode?: string;
  expiresIn?: string;
};

export function VendorSurveyOtp(props: VendorSurveyOtpProps) {
  const {
    displayName = "{{ .DisplayName }}",
    otpCode = "{{ .OTPCode }}",
    expiresIn = "{{ .ExpiresIn }}",
  } = props;

  return (
    <Base preview="Your verification code for the vendor assessment">
      <Container>
        <Header />
        <Box>
          <Text className="text-lg font-bold mt-6 mb-2">
            Verification Code
          </Text>
          <Text className="text-lg mb-2">Hello {displayName},</Text>
          <Text className="mb-4">
            Please use the following verification code to access the vendor
            assessment:
          </Text>
          <Text className="text-3xl font-bold text-center tracking-widest my-6 py-4 bg-gray-100 rounded">
            {otpCode}
          </Text>
          <Text className="text-sm text-gray-600 mb-4">
            This code will expire in {expiresIn}. If you did not request this
            code, you can safely ignore this email.
          </Text>
          <Text className="text-sm text-gray-600 mt-4">
            For security reasons, never share this code with anyone.
          </Text>
        </Box>
      </Container>
    </Base>
  );
}

export default VendorSurveyOtp;
