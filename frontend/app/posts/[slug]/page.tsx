// pages/blog/[slug].tsx
import { Tag } from "@/components/tag";
import { notFound } from "next/navigation";
import { Metadata } from "next";
import { DashboardTableOfContents } from "@/components/toc";
import { fetchAllPosts, fetchAllSlugs, getPostData } from "@/lib/posts";
import { getTableOfContents } from "@/lib/toc";

interface PostPageProps {
  params: Promise<{ slug: string }>;
}
export const dynamicParams = true

export default async function PostPage({ params }: PostPageProps) {
  const { slug } = await params;
  const post = await getPostData(slug);
  if (!post) {
    notFound();
  }
  const { content, frontmatter ,rawMdx } = post;
  const toc = await getTableOfContents(rawMdx);
  return (
    <article className="container py-6 prose dark:prose-invert max-w-6xl mx-auto relative lg:gap-10 lg:py-10 xl:grid xl:grid-cols-[1fr_300px]">
      <div className="mx-auto w-full min-w-0">
        <h1 className="mb-2">{frontmatter.title}</h1>
        <div className="flex gap-2 mb-2">
          {frontmatter.tags?.map((tag) => (
            <Tag tag={tag} key={tag} />
          ))}
        </div>
        {frontmatter.description && (
          <p className="text-xl mt-0 text-muted-foreground">{frontmatter.description}</p>
        )}
        <hr className="my-4" />
        {content}
      </div>
      <aside className="hidden text-sm xl:block">
        <div className="sticky top-16 -mt-10 max-h-[calc(var(--vh)-4rem)] overflow-y-auto pt-10">
          <DashboardTableOfContents toc={toc} />
        </div>
      </aside>
    </article>
  );
}

export async function generateStaticParams(): Promise<{ slug: string }[]> {
  // const posts = await fetchAllPosts();
  const slugs = await fetchAllSlugs();

  // // Fetch all posts concurrently
  // const posts = await Promise.all(
  //   slugs.map(async (slug) => {
  //     const postData = await getPostData(slug);
  //     return postData;
  //   })
  // );

  // return posts
  //   .map((post) => ({ slug: post!.slug, post: post }));
  return slugs.map(slug => { return { slug: slug} })
}

export async function generateMetadata({ params }: PostPageProps): Promise<Metadata> {
  const { slug } = await params;
  const post = await getPostData(slug);

  if (!post) {
    return {};
  }

  const { frontmatter } = post;
  const ogSearchParams = new URLSearchParams();
  ogSearchParams.set("title", frontmatter.title);
  console.log(ogSearchParams.toString());

  return {
    title: frontmatter.title,
    description: frontmatter.description,
    authors: { name: "Leonardo Pinho" },
    openGraph: {
      title: frontmatter.title,
      description: frontmatter.description,
      type: "article",
      url: "/blog/" + slug,
      // images: [
      //   {
      //     url: `/api/og?${ogSearchParams.toString()}`,
      //     width: 1200,
      //     height: 630,
      //     alt: frontmatter.title,
      //   },
      // ],
    },
    twitter: {
      card: "summary_large_image",
      title: frontmatter.title,
      description: frontmatter.description,
      images: [`/api/og?${ogSearchParams.toString()}`],
    },
  };
}
