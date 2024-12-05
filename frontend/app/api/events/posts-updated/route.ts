import { NextRequest, NextResponse } from "next/server";
import { OAuth2Client } from "google-auth-library";
import { revalidatePath } from "next/cache";

const client = new OAuth2Client();
type EventPostUpdatedRequest = {
  message: {
    data: string; // base64 of PostData
    message_id: string;
    publish_time: string;
    attributes: {
      eventType:
        | "POST_CREATED"
        | "CONTENT_UPDATED"
        | "POST_DELETED"
        | "META_UPDATED";
      slug: string;
    };
  };
  subscription: string;
};
type PostData = {
  title: string;
  tags: string[];
  created_at: string;
  description: string;
  slug: string;
};
export async function POST(req: NextRequest) {
  const authenticated = await validatePubSubAuth(req);
  if (!authenticated) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }
  const { message }: EventPostUpdatedRequest = await req.json();
  const { attributes, data } = message;
  const { eventType, slug } = attributes;
  // const postData: PostData = JSON.parse(Buffer.from(data, 'base64').toString());
  switch (eventType) {
    case "POST_DELETED":
      revalidatePath(`/posts/${slug}`);
      break;
    case "POST_CREATED":
      revalidatePath("/");
      break;
    case "CONTENT_UPDATED":
      revalidatePath(`/posts/${slug}`);
      break;
    case "META_UPDATED":
      revalidatePath("/");
      break;
    // default
    default:
      return NextResponse.json({ error: "Unknown eventType" },{ status: 400 });
  }
  return NextResponse.json({ message: "ok" }, { status: 200 });
  // return just 200
}

async function validatePubSubAuth(req: NextRequest) {
  //VALIDATE AUTH HEADER AND ENVIRONMENT GCP PUB SUB
  if (process.env.ENVIRONMENT === "dev") {
    return true;
  }
  const authHeader = req.headers.get("authorization");
  if (!authHeader) {
    return false;
  }
  const idToken = authHeader.split(" ")[1]; // Extract the token from "Bearer <token>"
  const audience = process.env.GCP_PROJECT_ID;
  try {
    await client.verifyIdToken({ idToken, audience });
  } catch (error) {
    return false;
  }
  return true
}
