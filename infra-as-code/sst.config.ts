/// <reference path="./.sst/platform/config.d.ts" />

import { execSync } from "child_process";

export default $config({
  app(input) {
    return {
      name: "cloudificando",
      removal: input?.stage === "production" ? "retain" : "remove",
      home: "aws",
      providers: {
        aws: {
          region: "us-east-1",
          defaultTags: {
            tags: {
              app: "cloudificando",
            },
          },
        },
        gcp: {
          version: "8.10.0",
          region: "us-east1",
          defaultLabels: {
            app: "cloudificando",
          },
          project: "cloudificando",
        },
        cloudflare: "5.44.0",
      },
    };
  },

  async run() {
    const common =await import("./sst.common");
    const backend = await import ("./sst.backend")
    const frontend = await import ("./sst.frontend")
  },
});