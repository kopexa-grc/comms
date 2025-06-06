/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

import { render } from "@react-email/components";
import { templates } from "./src";
import { createElement } from "react";
import * as path from "path";
import * as fs from "fs";

async function doRender() {
  const dirPath = path.join(process.cwd(), "../templates");

  fs.mkdirSync(dirPath, { recursive: true });

  for (const template of templates) {
    const { component, name } = template;
    const html = await render(createElement(component));
    const txt = await render(createElement(component), {
      plainText: true,
    });

    fs.writeFileSync(path.join(dirPath, `${name}.html`), html);
    fs.writeFileSync(path.join(dirPath, `${name}.txt`), txt);
  }
}

doRender().then();
