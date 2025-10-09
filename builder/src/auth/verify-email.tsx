/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Button, Heading, Section, Text } from "@react-email/components";
import { Base } from "../partials/base";
import { Container } from "../partials/container";
import { Header } from "../partials/header";
import { Box } from "../partials/box";

type VerifyEmailProps = {
  displayName?: string;
  URL?: string;
};

export function VerifyEmail(props: VerifyEmailProps) {
  const { displayName = "{{ .DisplayName}}", URL = "{{ .URL }}" } = props;

  return (
    <Base preview="Verify your email address">
      <Container>
        <Header />
        <Box>
          <Heading>Confirm your email address</Heading>
          <Text className="text-lg leading-snug mb-7">
            Hello {displayName},
          </Text>
          <Text className="text-lg leading-snug mb-7">
            Please confirm your email address by clicking the button below. Once
            verified, you’ll be able to securely access your Kopexa account.
          </Text>
        </Box>
        <Box>
          <Section className="flex justify-center items-center px-4 py-6">
            <Button
              href={URL}
              className="bg-blue-600 text-white px-6 py-3 rounded font-bold my-4 hover:bg-blue-700"
            >
              Verify Email Address
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
            If you didn't request this email, there's nothing to worry about,
            you can safely ignore it.
          </Text>
        </Box>
      </Container>
    </Base>
  );
}

VerifyEmail.PreviewProps = {
  URL: "https://app.kopexa.com/auth/verify-email?token=OilPEqoZpsQ1NOJE7YlHkjpjh_zFsQm-vY5DwV9hXWg",
};

export default VerifyEmail;
