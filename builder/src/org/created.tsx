/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Heading, Text } from "@react-email/components";
import { Base } from "../partials/base";
import { Container } from "../partials/container";
import { Header } from "../partials/header";
import { Box } from "../partials/box";

type OrgCreatedProps = {
  displayName?: string;
  orgName?: string;
};

export function OrgCreated(props: OrgCreatedProps) {
  const { displayName = "{{ .DisplayName }}", orgName = "{{ .OrgName }}" } =
    props;

  return (
    <Base preview="Your organization has been created">
      <Container>
        <Header />
        <Box>
          <Heading>Welcome to {orgName}!</Heading>
          <Text className="text-lg leading-snug mb-7">
            Hello {displayName},
          </Text>
          <Text className="text-lg leading-snug mb-7">
            Congratulations! Your organization has been successfully created.
            You're now ready to start building your team and managing your
            resources.
          </Text>
        </Box>
        <Box>
          <Text className="text-lg font-semibold">Next steps:</Text>
          <Text className="text-base">
            1. Invite team members to join your organization
          </Text>
          <Text className="text-base">
            2. Set up your organization's settings and preferences
          </Text>
          <Text className="text-base">
            3. Configure your first space and resources
          </Text>
        </Box>
        <Box className="text-start mt-7">
          <Text className="text-lg leading-snug">
            We're excited to have you on board and can't wait to see what you'll
            build with Kopexa!
          </Text>
        </Box>
      </Container>
    </Base>
  );
}

export default OrgCreated;
