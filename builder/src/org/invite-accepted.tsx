/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Text, Button } from "@react-email/components";
import { Base } from "../partials/base";
import { Container } from "../partials/container";
import { Header } from "../partials/header";
import { Box } from "../partials/box";

type InviteAcceptedProps = {
  displayName?: string;
  inviteeName?: string;
  organization?: string;
};

export function InviteAccepted(props: InviteAcceptedProps) {
  const {
    displayName = "{{ .DisplayName }}",
    inviteeName = "{{ .InviteeName }}",
    organization = "{{ .Organization }}",
  } = props;

  return (
    <Base
      preview={`${inviteeName} has accepted your invitation to join ${organization}`}
    >
      <Container>
        <Header />
        <Box>
          <Text className="text-lg font-bold mt-6 mb-2">Good news! 🎉</Text>
          <Text className="text-lg mb-2">
            {inviteeName} has accepted your invitation and joined the
            organization "{organization}" on Kopexa.
          </Text>
          <Text className="mb-4">
            You can now assign responsibilities, grant access to spaces, or
            share assessments.
          </Text>
          <Button
            href="https://app.kopexa.com/org/members"
            className="bg-blue-600 text-white px-6 py-3 rounded font-bold my-4 hover:bg-blue-700"
          >
            Manage team members
          </Button>
          <Text className="text-sm text-gray-600 mt-4">
            This email was sent to you because you invited {inviteeName} to join
            your organization.
          </Text>
        </Box>
      </Container>
    </Base>
  );
}

export default InviteAccepted;
