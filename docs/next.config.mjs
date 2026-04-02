import nextra from "nextra";

const withNextra = nextra({
  theme: "nextra-theme-docs",
  themeConfig: "./theme.config.tsx",
});

const isProd = process.env.NODE_ENV === "production";

export default withNextra({
  output: "export",
  basePath: isProd ? "/migrations" : "",
  images: {
    unoptimized: true,
  },
});
