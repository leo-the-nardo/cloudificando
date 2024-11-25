// eslint-disable-next-line @typescript-eslint/triple-slash-reference
/// <reference path="./.sst/platform/config.d.ts" />
export default $config({
  app(input) {
    return {
      name: "frontend",
      removal: input?.stage === "production" ? "retain" : "remove",
      home: "aws",
      providers: { 
        cloudflare: "5.42.0",
        aws: {
          region: "us-east-1",
          defaultTags: {
            tags: {
              "app": "cloudificando-sst-frontend"
            }
          },
        }
      },
    };
  },
  async run() {
    new sst.aws.Nextjs("CloudificandoFrontend", {
      domain: {
        name: "cloudificando.com",
        aliases: ["blog.cloudificando.com"],
        dns: sst.cloudflare.dns()
      }
    });
  },
});
