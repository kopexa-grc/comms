/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Hr as HrComponent } from "@react-email/components";
import { cn } from "../lib/cn";

type HrProps = {
  className?: string;
};

export const Hr = (props: HrProps) => {
  const { className } = props;

  return <HrComponent className={cn("my-5 border-slate-200", className)} />;
};
