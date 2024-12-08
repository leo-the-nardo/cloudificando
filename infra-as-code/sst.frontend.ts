import { pubsubTopic } from "./sst.common";
const cloudfrontSsmDistroIdPath = process.env.SST_FRONTEND_DISTROID_SSM_DEST!;
const frontend = new sst.aws.Nextjs("CloudificandoFrontend", {
  domain: {
    name: process.env.FRONTEND_PROD_DOMAIN!,
    aliases: ["blog." + process.env.FRONTEND_PROD_DOMAIN!],
    dns: sst.cloudflare.dns({proxy: true})
  },
  permissions: [
    {
      actions: ["ssm:GetParameter"],
      resources: [
      `arn:aws:ssm:${process.env.AWS_REGION}:${process.env.AWS_ACCOUNT_ID}:parameter${cloudfrontSsmDistroIdPath}`,
      ],
    }
  ],
  invalidation: {
    paths: "all"
  },
  environment: {
    NEW_RELIC_APP_NAME: process.env.NEW_RELIC_APP_NAME!,
    NEW_RELIC_LICENSE_KEY: process.env.NEW_RELIC_LICENSE_KEY!,
    REVALIDATE_SECRET: process.env.REVALIDATE_SECRET!,
    NEXT_PUBLIC_BACKEND_URL: process.env.NEXT_PUBLIC_BACKEND_URL!,
    PROD_DOMAIN: process.env.FRONTEND_PROD_DOMAIN!,
    CONTENT_GITHUB_REPO: process.env.CONTENT_GITHUB_REPO!,
    CONTENT_GITHUB_BRANCH: process.env.CONTENT_GITHUB_BRANCH!,
    CONTENT_GITHUB_MDX_PATH: process.env.CONTENT_GITHUB_MDX_PATH!,
    OTEL_EXPORTER_OTLP_ENDPOINT: process.env.OTEL_EXPORTER_OTLP_ENDPOINT!,
    OTEL_EXPORTER_OTLP_HEADERS: process.env.OTEL_EXPORTER_OTLP_HEADERS!,
    OTEL_RESOURCE_ATTRIBUTES: process.env.OTEL_RESOURCE_ATTRIBUTES!,
    ENVIRONMENT: process.env.ENVIRONMENT!,
    API_KEY: process.env.API_KEY!,
    SSM_CLOUDFRONT_DISTRIBUTION_ID_PATH: cloudfrontSsmDistroIdPath,
  },
  path: "../frontend"
});
const frontendSub = new gcp.pubsub.Subscription("posts-updated-subscription", {
  topic: pubsubTopic.name,
  ackDeadlineSeconds: 60,
  pushConfig: {
    pushEndpoint: "https://" + process.env.PROD_DOMAIN! + "/api/events/posts-updated",
    oidcToken: {
      serviceAccountEmail: process.env.GCP_SERVICE_ACCOUNT_EMAIL!,
    }
  },
  messageRetentionDuration: "86300s",
  retryPolicy: {
    minimumBackoff: "10s",
    maximumBackoff: "30s",
  },
});
frontend.nodes.cdn.nodes.distribution.id.apply((id) => {
  new aws.ssm.Parameter(
    cloudfrontSsmDistroIdPath,
    {
      name: cloudfrontSsmDistroIdPath,
      type: "String",
      value: id,
    },
  );
});

export {}