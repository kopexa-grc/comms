/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Text, Button } from "@react-email/components";
import { Base } from "../partials/base";
import { Container } from "../partials/container";
import { Header } from "../partials/header";
import { Box } from "../partials/box";

type VendorAssessmentRequestProps = {
  displayName?: string;
  actorName?: string;
  organizationName?: string;
  assessmentUrl?: string;
};

export function VendorAssessmentRequest(props: VendorAssessmentRequestProps) {
  const {
    displayName = "{{ .DisplayName }}",
    actorName = "{{ .ActorName }}",
    organizationName = "{{ .OrganizationName }}",
    assessmentUrl = "{{ .AssessmentUrl }}",
  } = props;

  return (
    <Base
      preview={`${organizationName} has requested a vendor assessment from you`}
    >
      <Container>
        <Header />
        <Box>
          <Text className="text-lg font-bold mt-6 mb-2">
            Vendor Assessment Request 📋
          </Text>
          <Text className="text-lg mb-2">Hello {displayName},</Text>
          <Text className="mb-4">
            {actorName} from {organizationName} has requested you to complete a
            vendor assessment. This assessment helps us ensure our vendors meet
            our quality and compliance standards.
          </Text>
          <Text className="mb-4">
            Please click the button below to start the assessment process. The
            assessment should take approximately 15-20 minutes to complete.
          </Text>
          <Button
            href={assessmentUrl}
            className="bg-blue-600 text-white px-6 py-3 rounded font-bold my-4 hover:bg-blue-700"
          >
            Start Assessment
          </Button>
          <Text className="text-sm text-gray-600 mb-2">
            If the button doesn't work, you can copy and paste this link into
            your browser:
          </Text>
          <Text className="text-sm text-gray-600 break-all mb-4">
            {assessmentUrl}
          </Text>
          <Text className="mb-4">
            If you have any questions or need assistance, please don't hesitate
            to contact our support team at{" "}
            <a href="mailto:support@kopexa.com" className="text-blue-600">
              support@kopexa.com
            </a>
          </Text>
          <Text className="text-sm text-gray-600 mt-4">
            This assessment is an important part of our vendor management
            process and helps us maintain high standards of quality and
            compliance.
          </Text>
        </Box>
      </Container>
    </Base>
  );
}

export default VendorAssessmentRequest;
