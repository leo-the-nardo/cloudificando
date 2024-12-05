// lib/posts.ts
import { compileMDX } from "next-mdx-remote/rsc";
import Image from "next/image";
import { Callout } from "@/components/callout";
import { MDXRemoteProps } from 'next-mdx-remote/rsc';
import rehypeSlug from 'rehype-slug';
import rehypeAutolinkHeadings from 'rehype-autolink-headings';
import rehypePrettyCode from 'rehype-pretty-code';

export interface Frontmatter {
  title: string;
  description?: string;
  published: boolean;
  tags?: string[];
  date: string;
}

export interface PostData {
  slug: string;
  content: React.ReactNode;
  frontmatter: Frontmatter;
  rawMdx: string;
}



export interface FetchPostsResponse {
  items: {
    title: string;
    tags: string[];
    created_at: string;
    description: string | null;
    slug: string;
  }[];
  nextCursor: string | null;
}
export async function fetchPosts(cursor?: string): Promise<FetchPostsResponse> {
  const params = new URLSearchParams();
  if (cursor) params.append('cursor', cursor);
  params.append('limit', '4'); // Adjust limit as needed

  const response = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/blog/posts?${params.toString()}`,{
    next: { tags: ['posts']},
  });
  if (!response.ok) {

    throw new Error('Failed to fetch posts');
  }

  const data = await response.json();
  return data;

}

export async function getPostData(slug: string): Promise<PostData | null> {
  const GITHUB_REPO = process.env.CONTENT_GITHUB_REPO
  const GITHUB_BRANCH = process.env.CONTENT_GITHUB_BRANCH
  const GITHUB_MDX_PATH = process.env.CONTENT_GITHUB_MDX_PATH
  const url = `https://raw.githubusercontent.com/${GITHUB_REPO}/refs/heads/${GITHUB_BRANCH}/${GITHUB_MDX_PATH}/${slug}.mdx`;
  const rawMdxResponse = await fetch(url, { next: { tags: ['posts', slug] } });
  if (!rawMdxResponse.ok) {
    return null;
  }

  const rawMdx = await rawMdxResponse.text();
  
  const { content, frontmatter } = await compileMDX<Frontmatter>({
    source: rawMdx,
    components: { Image, Callout },
    options: { ...commonMdxOptions },
  });

  if (!frontmatter.published) {
    return null;
  }
  return { slug, content, frontmatter, rawMdx };
}

export async function fetchAllPosts(): Promise<PostData[]> {
  try {
    const slugs = await fetchAllSlugs();

    // Fetch all posts concurrently
    const posts = await Promise.all(
      slugs.map(async (slug) => {
        const postData = await getPostData(slug);
        return postData;
      })
    );

    // Filter out any nulls (e.g., unpublished posts or fetch failures)
    const publishedPosts = posts.filter((post): post is PostData => post !== null);

    return publishedPosts;
  } catch (error) {
    console.error('Error fetching all posts:', error);
    return [];
  }
}

export async function fetchAllSlugs(): Promise<string[]> {
  const GITHUB_REPO = process.env.CONTENT_GITHUB_REPO
  const GITHUB_BRANCH = process.env.CONTENT_GITHUB_BRANCH
  const GITHUB_MDX_PATH = process.env.CONTENT_GITHUB_MDX_PATH

  const apiUrl = `https://api.github.com/repos/${GITHUB_REPO}/contents/${GITHUB_MDX_PATH}?ref=${GITHUB_BRANCH}`;

  const response = await fetch(apiUrl, {
    headers: {
      Accept: 'application/vnd.github.v3.raw',
      // If the repository is private or you need higher rate limits, include an Authorization header:
      'Authorization': `Bearer ${process.env.CONTENT_GITHUB_PAT}`,
    },
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch slugs: ${response.status} ${response.statusText}`);
  }

  const files = await response.json();

  // Extract slugs by removing the '.mdx' extension
  const slugs = files
    .filter((file: any) => file.type === 'file' && file.name.endsWith('.mdx'))
    .map((file: any) => file.name.replace(/\.mdx$/, ''));

  return slugs;
}

// const headings = []
export const commonMdxOptions: MDXRemoteProps["options"] = {
  mdxOptions: {
    rehypePlugins: [
      rehypeSlug,
      // [rehypeExtractHeadings, { rank: 2, headings }],
      [rehypePrettyCode,{theme: 'github-dark'},],
      [rehypeAutolinkHeadings,
        {
          behavior: 'wrap',
          properties: {className: ['subheading-anchor'], ariaLabel: 'Link to section'},
        },
      ],
    ],
    remarkPlugins: [],  
  },
  parseFrontmatter: true
}