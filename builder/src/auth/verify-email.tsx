/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Heading, Section, Text } from "@react-email/components";
import { Base } from "../partials/base";
import { Container } from "../partials/container";
import { Header } from "../partials/header";
import { Box } from "../partials/box";

type VerifyEmailProps = {
  displayName: string;
  code: string;
};

export function VerifyEmail(props: VerifyEmailProps) {
  const { displayName = "{{ .DisplayName}}", code = "{{ .Code }}" } = props;

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
            Your confirmation code is below - enter it in your open browser
            window and we'll help you get signed in.
          </Text>
        </Box>
        <Box>
          <Section className="flex justify-center items-center px-4 py-6 bg-slate-100 rounded-md ">
            <Text className=" text-3xl text-center align-middle font-bold">
              {code}
            </Text>
          </Section>
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

export default VerifyEmail;
