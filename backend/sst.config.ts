/// <reference path="./.sst/platform/config.d.ts" />

import { execSync } from "child_process";

export default $config({
  app(input) {
    return {
      name: "backend",
      removal: input?.stage === "production" ? "retain" : "remove",
      home: "aws",
      providers: { 
        cloudflare: "5.42.0",
        aws: {
          region: "us-east-1",
          defaultTags: {
            tags: {
              "app": "cloudificando-sst-backend"
            }
          },
        }
      },
    };
  },
  async run() {
    const dynamo = new sst.aws.Dynamo("BlogTable", {
      fields: {
        PK: "string",
        SK: "string",
        GSI1PK: "string",
        GSI1SK: "string",
      },
      primaryIndex: { hashKey: "PK", rangeKey: "SK" },
      globalIndexes: {
        GSI1: {
          hashKey: "GSI1PK",
          rangeKey: "GSI1SK",
          projection: "all",
        },
      },
    });
    
    const build = `GOOS=linux GOARCH=amd64 go build -o ./build/bootstrap . && cp otel-config.yaml ./build/`;
  
    execSync(build, {
      stdio: "inherit",
    });

    const lambda = new  sst.aws.Function("MyFunction", {
      runtime: "provided.al2023",
      handler: "bootstrap",
      bundle: "./build",
      url: true,
      link: [dynamo],
      layers: [
        "arn:aws:lambda:us-east-1:184161586896:layer:opentelemetry-collector-amd64-0_12_0:1",
        "arn:aws:lambda:us-east-1:753240598075:layer:LambdaAdapterLayerX86:23",
      ],
      environment: {
        AWS_DYNAMO_TABLE_NAME: dynamo.name,
        OTEL_EXPORTER_OTLP_ENDPOINT: 'http://localhost:4317',
        OTEL_EXPORTER_OTLP_PROTOCOL: "grpc",
        OTEL_RESOURCE_ATTRIBUTES: `service.name=${process.env.PROD_DOMAIN!},service.version=0.0.1,deployment.environment=production`,
        OPENTELEMETRY_COLLECTOR_CONFIG_URI: "/var/task/otel-config.yaml", 
        OTLP_CLOUDIFICANDO_TOKEN: process.env.OTLP_CLOUDIFICANDO_TOKEN!,
        OTLP_CLOUDIFICANDO_ENDPOINT: process.env.OTLP_CLOUDIFICANDO_ENDPOINT!,
      }
    });
    new sst.aws.Router("MyRouter", {
      domain: {
        name: process.env.PROD_DOMAIN!,
        dns: sst.cloudflare.dns()
      },
      routes: {
        "/*": lambda.url,
      }
    });
  },
});
