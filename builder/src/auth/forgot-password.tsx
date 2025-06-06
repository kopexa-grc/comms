/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Heading, Section, Text } from "@react-email/components";
import { Base } from "../partials/base";
import { Container } from "../partials/container";
import { Header } from "../partials/header";
import { Box } from "../partials/box";

type ForgotPasswordProps = {
  displayName?: string;
  code?: string;
};

export function ForgotPassword(props: ForgotPasswordProps) {
  const { displayName = "{{ .DisplayName }}", code = "{{ .Code }}" } = props;

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
            We received a request to reset your password. Use the code below to
            set up a new password for your account.
          </Text>
        </Box>
        <Box>
          <Section className="flex justify-center items-center px-4 py-6 bg-slate-100 rounded-md">
            <Text className="text-3xl text-center align-middle font-bold">
              {code}
            </Text>
          </Section>
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
