import { NextRequest, NextResponse } from "next/server";
import { revalidatePath } from "next/cache";
import { getPostData } from "../../../lib/posts";
import { invalidateCloudFrontPaths } from "@/lib/opennext";
type PostData = {
  title: string;
  tags: string[];
  created_at: string;
  description: string | null;
  slug: string;
}

export async function POST(req: NextRequest) {
  const authenticated = await validateApiToken(req);
  if (!authenticated) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }
  try {
    await performHardSync()
    revalidatePath("/", "layout")
    invalidateCloudFrontPaths(["/*"])
  }catch (e){
    return NextResponse.error()
  }
  return NextResponse.json({ message: "ok" }, { status: 200 });
  // return just 200
}

async function performHardSync(): Promise<void> {
  const GITHUB_REPO = process.env.CONTENT_GITHUB_REPO;
  const GITHUB_BRANCH = process.env.CONTENT_GITHUB_BRANCH;
  const GITHUB_MDX_PATH = process.env.CONTENT_GITHUB_MDX_PATH;

  // Step 1: Fetch all MDX files (list slugs)
  const postsUrl = `https://api.github.com/repos/${GITHUB_REPO}/contents/${GITHUB_MDX_PATH}?ref=${GITHUB_BRANCH}`;
  const postsResponse = await fetch(postsUrl);
  if (!postsResponse.ok) {
    throw new Error('Failed to fetch posts directory.');
  }

  const files = await postsResponse.json();
  const slugs = files
    .filter((file: { type: string, name: string }) => file.type === 'file' && file.name.endsWith('.mdx'))
    .map((file: { name: string }) => file.name.replace('.mdx', ''));

  const posts: PostData[] = [];

  for (const slug of slugs) {
    const postData = await getPostData(slug);
    if (postData) {
      const { frontmatter } = postData;
      posts.push({
        title: frontmatter.title,
        tags: frontmatter.tags || [],
        created_at: frontmatter.date,
        description: frontmatter.description!,
        slug: slug,
      });
    }
  }

  // Step 3: Make the POST request with the data
  console.log("posts", posts)
  const syncResponse = await fetch(process.env.NEXT_PUBLIC_BACKEND_URL + "/blog/hardsync", {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({posts}),
  });

  if (!syncResponse.ok) {
    const errorDetails = await syncResponse.text();
    console.error(syncResponse.status, errorDetails)
    throw new Error(`Failed to sync posts: ${syncResponse.status} ${errorDetails}`);
  }

  console.log('Posts successfully synced.');
}

async function validateApiToken(req: NextRequest) {
  if (process.env.ENVIRONMENT === "dev") {
    return true;
  }
  const authHeader = req.headers.get("api-key");
  if (!authHeader) {
    return false;
  }
  const expectedToken = process.env.API_KEY;
  if (authHeader != expectedToken) {
    return false
  } 
  return true
}
