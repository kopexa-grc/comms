/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Text, Button } from "@react-email/components";
import { Base } from "../partials/base";
import { Container } from "../partials/container";
import { Header } from "../partials/header";
import { Box } from "../partials/box";

type InviteProps = {
  actor?: string;
  actorEmail?: string;
  organization?: string;
  message?: string;
  url?: string;
};

export function OrgInvite(props: InviteProps) {
  const {
    actor = "{{ .Actor }}",
    actorEmail = "{{ .ActorEmail }}",
    organization = "{{ .Organization }}",
    message = "{{ .Message }}",
    url = "{{ .Url }}",
  } = props;

  return (
    <Base preview={`Welcome to ${organization} on Kopexa!`}>
      <Container>
        <Header />
        <Box>
          <Text className="text-lg font-bold mt-6 mb-2">
            Welcome aboard! 🎉
          </Text>
          <Text className="text-lg mb-2">
            {actor} has invited you to join {organization} on Kopexa
          </Text>
          <Text className="mb-4">
            We're excited to have you join us! {actor} ({actorEmail}) has
            invited you to be part of {organization} on Kopexa. Together, we'll
            help streamline your compliance management, making audits, actions,
            and assessments more efficient than ever.
          </Text>
          <Text className="mb-4 italic">{message}</Text>
          <Button
            href={url}
            className="bg-blue-600 text-white px-6 py-3 rounded font-bold my-4 hover:bg-blue-700"
          >
            Accept Invitation
          </Button>
          <Text className="mb-2 text-gray-600">
            Or copy this link into your browser:
          </Text>
          <Text className="break-all mb-4 text-gray-600">{url}</Text>
        </Box>
      </Container>
    </Base>
  );
}

export default OrgInvite;
