/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Heading, Text } from "@react-email/components";
import { Base } from "../partials/base";
import { Container } from "../partials/container";
import { Header } from "../partials/header";
import { Box } from "../partials/box";

type WelcomeProps = {
  displayName?: string;
};

export function Welcome(props: WelcomeProps) {
  const { displayName = "{{ .DisplayName }}" } = props;

  return (
    <Base preview="Welcome to Kopexa">
      <Container>
        <Header />
        <Box>
          <Heading>Welcome to Kopexa!</Heading>
          <Text className="text-lg leading-snug mb-7">
            Hello {displayName},
          </Text>
          <Text className="text-lg leading-snug mb-7">
            We're thrilled to have you on board! Your email has been
            successfully verified, and your account is now ready to go.
          </Text>
          <Text className="text-lg leading-snug mb-7">
            You're now part of our community, and we're excited to help you make
            the most of your Kopexa experience. Whether you're here to
            streamline your workflow, enhance your productivity, or explore new
            possibilities, we're here to support your journey.
          </Text>
          <Text className="text-lg leading-snug">Welcome to the team!</Text>
        </Box>
      </Container>
    </Base>
  );
}

export default Welcome;
