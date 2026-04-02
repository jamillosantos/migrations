import React from "react";
import { DocsThemeConfig } from "nextra-theme-docs";

const config: DocsThemeConfig = {
  logo: <strong>migrations</strong>,
  project: {
    link: "https://github.com/jamillosantos/migrations",
  },
  docsRepositoryBase:
    "https://github.com/jamillosantos/migrations/tree/main/docs",
  footer: {
    text: "migrations - A flexible, driver-agnostic migration library for Go.",
  },
  useNextSeoProps() {
    return {
      titleTemplate: "%s - migrations",
    };
  },
  head: (
    <>
      <meta name="viewport" content="width=device-width, initial-scale=1.0" />
      <meta
        name="description"
        content="A flexible, driver-agnostic migration library for Go"
      />
    </>
  ),
};

export default config;
