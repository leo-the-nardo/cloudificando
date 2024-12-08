const messageSchema = `
{
  "type": "record",
  "name": "Avro",
  "fields": [
    { "name": "title", "type": "string" },
    { "name": "tags", "type": { "type": "array", "items": "string" } },
    { "name": "created_at", "type": "string" },
    { "name": "description", "type": "string" },
    { "name": "slug", "type": "string" }
  ]
}
`;
export const pubsubSchema = new gcp.pubsub.Schema("post-updated-schema", {
  definition: messageSchema,
  name: "post-updated-schema",
  type: "AVRO"
})
export const pubsubTopic = new gcp.pubsub.Topic("posts-updated-topic", {
  name: "posts-updated-topic",
  messageRetentionDuration: "86300s",
  schemaSettings: {
    encoding: "JSON",
    schema: pubsubSchema.id
  }
 });
export const ssmPubsubTopic = new aws.ssm.Parameter("pubsub-topic-name", {
  name: process.env.SST_COMMON_PUBSUB_TOPIC_NAME_SSM_DEST!,
  type: "String",
  value: pubsubTopic.id,
})
const ROOT_DOMAIN = process.env.ROOT_DOMAIN!
export const zone = (await cloudflare.getZone({name: ROOT_DOMAIN}))
export const cacheBypass = new cloudflare.Ruleset("cloudflare-cache-bypass", {
  phase: "http_request_cache_settings",
  description: "Bypass cloudflar ecache for the blog api due cloudfront handles caches",
  name: "disable-cache-on-cloudfront-origins",
  kind: "zone",
  zoneId: zone.id,
  rules: [
    {
      action: "set_cache_settings",
      actionParameters: {
        cache: false,
      },
      enabled: true,
      expression: `http.request.full_uri wildcard "https://api.${ROOT_DOMAIN}/*" or http.request.full_uri wildcard "https://${ROOT_DOMAIN}/*" or http.request.full_uri wildcard "https://blog.${ROOT_DOMAIN}/*"`,
    },
  ],
});

export {}