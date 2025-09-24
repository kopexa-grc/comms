/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Text, Button } from "@react-email/components";
import { Base } from "../partials/base";
import { Container } from "../partials/container";
import { Header } from "../partials/header";
import { Box } from "../partials/box";

type ReviewOverdueProps = {
  entity?: string;
  entityName?: string;
  space?: string;
  url?: string;
};

export function ReviewOverdue(props: ReviewOverdueProps) {
  const {
    entity = "{{ .Entity }}",
    entityName = "{{ .EntityName }}",
    space = "{{ .Space }}",
    url = "{{ .Url }}",
  } = props;

  return (
    <Base preview={`${entity}: ${entityName} is overdue for review`}>
      <Container>
        <Header />
        <Box>
          <Text className="text-lg font-bold mt-6 mb-2">
            Action required ⚠️
          </Text>
          <Text className="text-lg mb-2">
            The {entity} <strong>{entityName}</strong> in{" "}
            <strong>{space}</strong> is overdue for review.
          </Text>
          <Text className="mb-4">
            Please take a moment to review this entity and complete the
            outstanding check.
          </Text>
          <Button
            href={url}
            className="bg-blue-600 text-white px-6 py-3 rounded font-bold my-4 hover:bg-blue-700"
          >
            Review Now
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

export default ReviewOverdue;
