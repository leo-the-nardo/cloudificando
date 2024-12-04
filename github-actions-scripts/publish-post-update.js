const fs = require('fs');
const path = require('path');
const { PubSub } = require('@google-cloud/pubsub');
const matter = require('gray-matter');
const { execSync } = require('child_process');

// Configure Pub/Sub
const topicName = process.env.GCP_TOPIC_NAME;

// Initialize PubSub client using default credentials
const pubSubClient = new PubSub();

// Extract frontmatter from an MDX file
function extractFrontmatter(filePath) {
  const content = fs.readFileSync(filePath, 'utf-8');
  const { data: frontmatter } = matter(content);
  return frontmatter;
}

// Determine event type based on git diff
function determineEventType(filePath) {
  const slug = path.basename(filePath, '.mdx');
  try {
    const creationOutput = execSync(`git log --diff-filter=A -- ${filePath}`).toString().trim();
    if (creationOutput) {
      return 'POST_CREATED';
    }

    if (!fs.existsSync(filePath)) {
      return 'POST_DELETED';
    }

    const diffOutput = execSync(`git diff HEAD~1 HEAD -- ${filePath}`).toString().trim();
    if (diffOutput.includes('title') || diffOutput.includes('tags') || diffOutput.includes('description')) {
      return 'META_UPDATED';
    }

    return 'CONTENT_UPDATED';
  } catch (error) {
    console.error(`Error determining event type for ${slug}:`, error);
    throw error;
  }
}

// Publish event to Pub/Sub
async function publishEvent(slug, frontmatter, eventType) {
  const post = {
    title: frontmatter.title,
    tags: frontmatter.tags,
    created_at: frontmatter.date,
    description: frontmatter.description,
    slug,
  };

  const message = {
    data: post,
    attributes: {
      eventType,
      slug,
    },
  };

  // const dataBuffer = Buffer.from(JSON.stringify({ message}));

  try {
    pubSubClient.topic(topicName).publishMessage({data: JSON.stringify(message.data) , attributes: message.attributes})
    console.log(`Message published for slug: ${slug}`);
  } catch (error) {
    console.log(`Failed to publish message for slug: ${slug}`, error);
    console.error(`Failed to publish message for slug: ${slug}`, error);
    process.exit(1);
  }
}

// Main script
(async () => {
  const changedFiles = process.env.CHANGED_FILES.split(',').filter((file) => file.endsWith('.mdx'));
  console.log("changed files:",changedFiles)
  for (const file of changedFiles) {
    const slug = path.basename(file, '.mdx');
    const frontmatter = extractFrontmatter(file);
    const eventType = determineEventType(file);

    console.log(`Detected change in ${file}: Event Type - ${eventType}`);
    await publishEvent(slug, frontmatter, eventType);
  }
  console.log("Published.")
})();
