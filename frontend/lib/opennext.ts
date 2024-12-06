import { CloudFrontClient, CreateInvalidationCommand } from "@aws-sdk/client-cloudfront";
import { GetParameterCommand, SSMClient } from "@aws-sdk/client-ssm";
// https://opennext.js.org/aws/inner_workings/caching#cloudfront-cache-invalidation 
const region = process.env.AWS_REGION;
const cloudFront = new CloudFrontClient({
  region: region,
});
const ssm = new SSMClient({
  region: region,
});

export async function invalidateCloudFrontPaths(paths: string[]) {
  if (process.env.ENVIRONMENT !== "production") {
    return;
  }
  const distributionId = (await ssm.send(
    new GetParameterCommand({
      Name: process.env.SSM_CLOUDFRONT_DISTRIBUTION_ID_PATH,
    }),
  )).Parameter?.Value
  await cloudFront.send(
    new CreateInvalidationCommand({
      // Set CloudFront distribution ID here
      DistributionId: distributionId,
      InvalidationBatch: {
        CallerReference: `${Date.now()}`,
        Paths: {
          Quantity: paths.length,
          Items: paths,
        },
      },
    }),
  );
}