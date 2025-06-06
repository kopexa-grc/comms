/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Container as ContainerComponent } from "@react-email/components";

type ContainerProps = {
  children: React.ReactNode;
};

export const Container = (props: ContainerProps) => {
  const { children } = props;

  return (
    <ContainerComponent className="bg-white mx-auto pt-[20px] pb-[40px]">
      {children}
    </ContainerComponent>
  );
};
