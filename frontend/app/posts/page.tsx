// /app/posts/page.tsx

import { Metadata } from "next";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { TagList} from "@/components/tag";
import InfinitePosts from "@/components/infinite-posts";
import { Suspense } from "react";
export const metadata: Metadata = {
  title: "Cloudificando",
  description: "This is a description",
};
type PageProps = {
  searchParams: Promise<{
    tag?: string;
  }>;
};
async function BlogPage() {
  return (
    <div className="container max-w-4xl py-6 lg:py-10">
      <div className="flex-1 space-y-4 items-start gap-4 md:flex-row md:justify-between md:gap-8">
        <h1 className="inline-block font-black text-4xl lg:text-5xl">Blog</h1>
        <p className="text-xl text-muted-foreground">
          My ramblings on all things web dev.
        </p>
      </div>
      <div className="grid grid-cols-12 gap-3 mt-8">
        <div className="col-span-12 col-start-1 sm:col-span-8">
          <hr />
          <Suspense>
          <InfinitePosts />
          </Suspense>
        </div>

        <Card className="col-span-12 row-start-3 h-fit sm:col-span-4 sm:col-start-9 sm:row-start-1">
          <CardHeader>
            <CardTitle>Tags</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-wrap gap-2">
          <Suspense>
            <TagList />
          </Suspense>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default BlogPage;
