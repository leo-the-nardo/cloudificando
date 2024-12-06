const fs = require('fs');
const path = require('path');
const { PubSub } = require('@google-cloud/pubsub');
const matter = require('gray-matter');
const { execSync } = require('child_process');

// Configure Pub/Sub
const topicName = process.env.GCP_TOPIC_NAME;
const pubSubClient = new PubSub();

// Extract frontmatter from an MDX file
function extractFrontmatter(filePath) {
  const content = fs.readFileSync(filePath, 'utf-8');
  const { data: frontmatter } = matter(content);
  return frontmatter;
}

// Determine event types based on git diff
function determineEventTypes(filePath) {
  try {
    // Check if file was created
    const createdFiles = execSync(`git diff --diff-filter=A --name-only HEAD~1`).toString().trim().split('\n');
    if (createdFiles.includes(filePath)) {
      return ['POST_CREATED'];
    }

    // Check if file was deleted
    const deletedFiles = execSync(`git diff --diff-filter=D --name-only HEAD~1`).toString().trim().split('\n');
    if (deletedFiles.includes(filePath)) {
      return ['POST_DELETED'];
    }

    // Check for content and metadata changes
    const diffOutput = execSync(`git diff HEAD~1 HEAD -- ${filePath}`).toString();
    const hasMetadataChanges = ['title', 'tags', 'description', 'date'].some((field) => diffOutput.includes(field));
    const hasContentChanges = diffOutput.replace(/---\n[\s\S]*?---\n/, '').trim() !== '';

    const events = [];
    if (hasMetadataChanges) events.push('META_UPDATED');
    if (hasContentChanges) events.push('CONTENT_UPDATED');

    return events.length > 0 ? events : ['UNKNOWN'];
  } catch (error) {
    console.error(`Error determining event types for ${filePath}:`, error);
    throw error;
  }
}

// Publish event to Pub/Sub
async function publishEvent(slug, frontmatter, eventType) {
  const post = {
    title: frontmatter.title || null,
    tags: frontmatter.tags || [],
    created_at: frontmatter.date || null,
    description: frontmatter.description || null,
    slug,
  };

  const message = {
    data: post,
    attributes: {
      eventType,
      slug,
    },
  };

  try {
    await pubSubClient
      .topic(topicName)
      .publishMessage({ data: Buffer.from(JSON.stringify(message.data)), attributes: message.attributes });
    console.log(`Message published for slug: ${slug}, Event Type: ${eventType}`);
  } catch (error) {
    console.error(`Failed to publish message for slug: ${slug}, Event Type: ${eventType}`, error);
    process.exit(1);
  }
}

// Main script
(async () => {
  const changedFiles = process.env.CHANGED_FILES.split(',').filter((file) => file.endsWith('.mdx'));
  console.log("Changed files:", changedFiles);

  for (const file of changedFiles) {
    if (!fs.existsSync(file)) {
      console.log(`File does not exist: ${file}`);
      continue;
    }
    const slug = path.basename(file, '.mdx');
    const frontmatter = extractFrontmatter(file);
    const eventTypes = determineEventTypes(file);

    console.log(`Detected changes in ${file}: Event Types - ${eventTypes}`);
    for (const eventType of eventTypes) {
      await publishEvent(slug, frontmatter, eventType);
    }
  }
  console.log("All events published.");
})();
