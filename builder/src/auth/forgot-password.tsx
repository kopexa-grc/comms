/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Button, Heading, Section, Text } from "@react-email/components";
import { Base } from "../partials/base";
import { Container } from "../partials/container";
import { Header } from "../partials/header";
import { Box } from "../partials/box";

type ForgotPasswordProps = {
  displayName?: string;
  URL?: string;
};

export function ForgotPassword(props: ForgotPasswordProps) {
  const { displayName = "{{ .DisplayName }}", URL = "{{ .URL }}" } = props;

  return (
    <Base preview="Reset your password">
      <Container>
        <Header />
        <Box>
          <Heading>Reset your password</Heading>
          <Text className="text-lg leading-snug mb-7">
            Hello {displayName},
          </Text>
          <Text className="text-lg leading-snug mb-7">
            We received a request to reset your password. Click the button below
            to choose a new one for your account.
          </Text>
        </Box>
        <Box>
          <Section className="flex justify-center items-center px-4 py-6">
            <Button
              href={URL}
              className="bg-[#13274E] text-white text-base font-medium px-6 py-3 rounded-md no-underline"
            >
              Reset Password
            </Button>
          </Section>
        </Box>

        <Box className="text-start mt-4">
          <Text className="text-sm leading-tight">
            Or copy and paste this link into your browser:
          </Text>
          <Text className="text-xs text-slate-600 break-all mt-2">{URL}</Text>
        </Box>

        <Box className="text-start">
          <Text className="text-sm leading-tight">
            If you didn't request this password reset, you can safely ignore
            this email. Your password will remain unchanged.
          </Text>
        </Box>
      </Container>
    </Base>
  );
}

export default ForgotPassword;
