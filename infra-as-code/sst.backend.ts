import { execSync } from "child_process";
import { pubsubTopic } from "./sst.common";
const dynamo = sst.aws.Dynamo.get("BlogTable", "BlogTable"); // go app migrations manages it
const build = `GOOS=linux GOARCH=amd64 go build -o ./build/bootstrap . && cp otel-config.yaml ./build/`;
execSync(build, {
  stdio: "inherit",
  cwd: "../backend"
});
const backend = new sst.aws.Function("MyFunction", {
  runtime: "provided.al2023",
  handler: "bootstrap",
  bundle: "../backend/build",
  url: true,
  permissions: [
    {
      actions: ["ssm:GetParameter"],
      resources: [
        "arn:aws:ssm:::parameter/cloudificando/backend/cloudfront/distribution-id",
      ],
    },
  ],
  link: [dynamo, pubsubTopic],
  layers: [
    "arn:aws:lambda:us-east-1:184161586896:layer:opentelemetry-collector-amd64-0_12_0:1",
    "arn:aws:lambda:us-east-1:753240598075:layer:LambdaAdapterLayerX86:23",
  ],
  architecture: "x86_64",
  memory: "128 MB",
  // logging: false,
  environment: {
    AWS_DYNAMO_TABLE_NAME: dynamo.name,
    OTEL_EXPORTER_OTLP_ENDPOINT: "http://localhost:4317",
    OTEL_EXPORTER_OTLP_PROTOCOL: "grpc",
    OTEL_RESOURCE_ATTRIBUTES: `service.name=${process.env
      .BACKEND_PROD_DOMAIN!},service.version=0.0.1,deployment.environment=production`,
    OPENTELEMETRY_COLLECTOR_CONFIG_URI: "/var/task/otel-config.yaml",
    OTLP_CLOUDIFICANDO_TOKEN: process.env.OTLP_CLOUDIFICANDO_TOKEN!,
    OTLP_CLOUDIFICANDO_ENDPOINT: process.env.OTLP_CLOUDIFICANDO_ENDPOINT!,
    PROD_DOMAIN: process.env.BACKEND_PROD_DOMAIN!,
    ALLOWED_ORIGINS: process.env.BACKEND_ALLOWED_ORIGINS!,
    AWS_SSM_CLOUDFRONT_DISTRO_ID_PATH:
      "/cloudificando/backend/cloudfront/distribution-id",
    GCP_PROJECT_ID: "cloudificando",
    ENVIROMENT: "production",
    GIN_MODE: "release",
  },
});
const backendCloudfront = new sst.aws.Router("MyRouter", {
  domain: {
    name: process.env.BACKEND_PROD_DOMAIN!,
    //@ts-ignore
    dns: sst.cloudflare.dns({proxy: true}),
  },
  routes: {
    "/*": backend.url,
  },
  invalidation: true,
  
});
const backendSub = new gcp.pubsub.Subscription(
  "posts-updates-subscription",
  {
    topic: pubsubTopic.name,
    ackDeadlineSeconds: 60,
    pushConfig: {
      pushEndpoint:
        "https://" + process.env.BACKEND_PROD_DOMAIN! + "/blog/posts-updated",
      oidcToken: {
        serviceAccountEmail: process.env.GCP_SERVICE_ACCOUNT_EMAIL!,
      },
    },
    messageRetentionDuration: "86300s",
    retryPolicy: {
      minimumBackoff: "10s",
      maximumBackoff: "30s",
    },
  },
);
backendCloudfront.nodes.cdn.nodes.distribution.id.apply((id) => {
  new aws.ssm.Parameter(
    "/cloudificando/backend/cloudfront/distribution-id",
    {
      name: "/cloudificando/backend/cloudfront/distribution-id",
      type: "String",
      value: id,
    },
  );
});

export {};

