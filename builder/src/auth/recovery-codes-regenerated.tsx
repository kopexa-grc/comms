/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Button, Heading, Section, Text } from "@react-email/components";
import { Base } from "../partials/base";
import { Container } from "../partials/container";
import { Header } from "../partials/header";
import { Box } from "../partials/box";

type RecoveryCodesRegeneratedProps = {
  displayName?: string;
  /** Link to security settings or recovery-codes page */
  URL?: string;
};

export function RecoveryCodesRegenerated(props: RecoveryCodesRegeneratedProps) {
  const { displayName = "{{ .DisplayName }}", URL = "{{ .URL }}" } = props;

  return (
    <Base preview="Your two-factor recovery codes were re-generated">
      <Container>
        <Header />

        <Box>
          <Heading>Your two-factor recovery codes were re-generated</Heading>

          <Text className="text-lg leading-snug mb-7">
            Hello {displayName},
          </Text>

          <Text className="text-lg leading-snug mb-7">
            We’re letting you know that your two-factor authentication (2FA)
            recovery codes were re-generated. For security reasons, we do not
            include the codes in this email.
          </Text>

          <Text className="text-base leading-snug mb-6">
            Please store your new recovery codes in a safe place. You can review
            or download them from your account’s security settings:
          </Text>
        </Box>

        <Box>
          <Section className="flex justify-center items-center px-4 py-6">
            <Button
              href={URL}
              className="bg-[#13274E] text-white text-base font-medium px-6 py-3 rounded-md no-underline"
            >
              Review Security Settings
            </Button>
          </Section>
        </Box>

        <Box className="text-start mt-6">
          <Text className="text-sm leading-tight">
            If this wasn’t you, we recommend resetting your password and
            reviewing recent activity.
          </Text>
          <Text className="text-sm leading-tight mt-2">
            For your security, recovery codes are only shown after you sign in.
          </Text>
          <Text className="text-xs text-slate-600 mt-4">
            If you didn’t request this change, please contact support
            immediately.
          </Text>
        </Box>
      </Container>
    </Base>
  );
}

export default RecoveryCodesRegenerated;
