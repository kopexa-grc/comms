/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Button, Heading, Text } from "@react-email/components";
import { Base } from "./partials/base";
import { Box } from "./partials/box";
import { Container } from "./partials/container";
import { Header } from "./partials/header";

type SubscribeProps = {
  url: string;
};

export function Subscribe(props: SubscribeProps) {
  const { url = "{{ .URL }}" } = props;
  return (
    <Base preview="Please verify your email to complete the subscription process">
      <Container>
        <Header />
        <Box>
          <Heading>Thank you for subscribing!</Heading>
          <Text>
            Thank you for subscribing to Kopexa - to complete the subscription
            process, please verify your email address by clicking the following
            link:
          </Text>
          <Button
            href={url}
            className="bg-blue-600 text-white px-6 py-3 rounded font-bold my-4 hover:bg-blue-700"
          >
            Confirm Subscription
          </Button>
          <Text className="text-sm text-gray-600 mb-2">
            If the button doesn't work, you can copy and paste this link into
            your browser:
          </Text>
          <Text className="text-sm text-gray-600 break-all mb-4">{url}</Text>
          <Text className="mb-4">
            If you have any questions or need assistance, please don't hesitate
            to contact our support team at{" "}
            <a href="mailto:support@kopexa.com" className="text-blue-600">
              support@kopexa.com
            </a>
          </Text>
        </Box>
      </Container>
    </Base>
  );
}

export default Subscribe;
