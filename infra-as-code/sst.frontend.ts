import { pubsubTopic } from "./sst.common";

const frontend = new sst.aws.Nextjs("CloudificandoFrontend", {
  domain: {
    name: process.env.FRONTEND_PROD_DOMAIN!,
    aliases: ["blog." + process.env.FRONTEND_PROD_DOMAIN!],
    //@ts-ignore
    dns: sst.cloudflare.dns({proxy: true})
  },
  invalidation: {
    paths: "all"
  },
  environment: {
    NEW_RELIC_APP_NAME: process.env.NEW_RELIC_APP_NAME!,
    NEW_RELIC_LICENSE_KEY: process.env.NEW_RELIC_LICENSE_KEY!,
    REVALIDATE_SECRET: process.env.REVALIDATE_SECRET!,
    OTEL_TOKEN: process.env.OTEL_TOKEN!,
    NEXT_PUBLIC_BACKEND_URL: process.env.NEXT_PUBLIC_BACKEND_URL!,
    PROD_DOMAIN: process.env.FRONTEND_PROD_DOMAIN!,
    GPC_TOPIC_ID: process.env.GPC_TOPIC_ID!,
    GCP_TOPIC_NAME: process.env.GCP_TOPIC_NAME!,
    GCP_SERVICE_ACCOUNT_EMAIL: process.env.GCP_SERVICE_ACCOUNT_EMAIL!,
    CONTENT_GITHUB_REPO: process.env.CONTENT_GITHUB_REPO!,
    CONTENT_GITHUB_BRANCH: process.env.CONTENT_GITHUB_BRANCH!,
    CONTENT_GITHUB_MDX_PATH: process.env.CONTENT_GITHUB_MDX_PATH!,
    OTEL_EXPORTER_OTLP_ENDPOINT: process.env.OTEL_EXPORTER_OTLP_ENDPOINT!,
    OTEL_EXPORTER_OTLP_PROTOCOL: process.env.OTEL_EXPORTER_OTLP_PROTOCOL!,
    OTEL_RESOURCE_ATTRIBUTES: process.env.OTEL_RESOURCE_ATTRIBUTES!,
    ENVIROMENT: process.env.ENVIRONMENT!,
    OTEL_EXPORTER_OTLP_HEADERS: process.env.OTEL_EXPORTER_OTLP_HEADERS!
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
export {}