/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Text, Button } from "@react-email/components";
import { Base } from "../partials/base";
import { Container } from "../partials/container";
import { Header } from "../partials/header";
import { Box } from "../partials/box";

type PasswordResetSuccessProps = {
  displayName?: string;
};

export function PasswordResetSuccess(props: PasswordResetSuccessProps) {
  const { displayName = "{{ .DisplayName }}" } = props;

  return (
    <Base preview="Your password has been successfully reset">
      <Container>
        <Header />
        <Box>
          <Text className="text-lg font-bold mt-6 mb-2">
            Password Reset Successful! 🔒
          </Text>
          <Text className="text-lg mb-2">Hello {displayName},</Text>
          <Text className="mb-4">
            Your password has been successfully reset. Your account is now
            secure and ready to use with your new password.
          </Text>
          <Text className="mb-4">
            If you did not make this change, please contact our support team
            immediately at{" "}
            <a href="mailto:support@kopexa.com" className="text-blue-600">
              support@kopexa.com
            </a>
          </Text>
          <Button
            href="https://app.kopexa.com/auth/login"
            className="bg-blue-600 text-white px-6 py-3 rounded font-bold my-4 hover:bg-blue-700"
          >
            Sign in to your account
          </Button>
          <Text className="text-sm text-gray-600 mt-4">
            For security reasons, this email was sent to all email addresses
            associated with your account.
          </Text>
        </Box>
      </Container>
    </Base>
  );
}

export default PasswordResetSuccess;
