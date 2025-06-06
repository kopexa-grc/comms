/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { Body, Head, Html, Preview, Tailwind } from "@react-email/components";
import { Show } from "../lib/when";
import { Footer } from "./footer";

type BaseProps = {
  children: React.ReactNode;
  preview: string;
};

export function Base(props: BaseProps) {
  const { children, preview } = props;

  return (
    <Html>
      <Head />
      <Show when={preview}>
        <Preview>{preview}</Preview>
      </Show>
      <Tailwind>
        <Body className="bg-slate-100 text-slate-900 font-sans">
          {children}

          <Footer />
        </Body>
      </Tailwind>
    </Html>
  );
}
