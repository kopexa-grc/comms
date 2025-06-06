/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Section } from "@react-email/components";
import { cn } from "../lib/cn";

type BoxProps = {
  children: React.ReactNode;
  className?: string;
};

export const Box = (props: BoxProps) => {
  const { children, className } = props;

  return <Section className={cn("px-[48px]", className)}>{children}</Section>;
};
