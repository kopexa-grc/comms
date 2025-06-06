/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Box } from "./box";
import { Container, Link, Text } from "@react-email/components";
import { Hr } from "./hr";

export const Footer = () => {
  return (
    <Container className="mx-auto pt-[20px] pb-[40px]">
      <Box>
        <Text>
          You are receiving this email because your organization is using Kopexa
          or has interacted with our services. This message may contain
          important updates regarding your compliance activities, organization
          settings, or legal obligations.
        </Text>
        <Text>
          If you have any questions or need assistance, please contact our
          support team at{" "}
          <Link href="mailto:support@kopexa.com">support@kopexa.com</Link>.
        </Text>
        <Text>
          This email was sent by Kopexa, a platform for managing compliance and
          organization settings.
        </Text>
        <Hr />
        <Text>
          Kopexa GmbH
          <br />
          Unterjörn 19a
          <br />
          24536 Neumünster
          <br />
          Germany
        </Text>
      </Box>
    </Container>
  );
};
