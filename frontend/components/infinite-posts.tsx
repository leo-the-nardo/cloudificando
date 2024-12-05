// InfinitePosts.tsx

"use client";

import React from 'react';
import {
  useInfiniteQuery,
  type QueryFunctionContext,
  type InfiniteData,
} from '@tanstack/react-query';
import InfiniteScroll from 'react-infinite-scroll-component';
import { PostItem } from "@/components/post-item";
import { useSearchParams } from 'next/navigation';

interface Post {
  slug: string;
  created_at: string;
  title: string;
  description: string;
  tags: string[];
}

interface ApiResponse {
  items: Post[];
  nextCursor: string | null;
}

type PostsQueryKey = readonly ['posts', string];

const fetchPosts = async ({
  pageParam = undefined,
  queryKey,
}: QueryFunctionContext<PostsQueryKey, string | undefined>): Promise<ApiResponse> => {
  const [_key, tag] = queryKey;
  const params = new URLSearchParams();

  if (pageParam) params.append('cursor', pageParam as string);
  if (tag) params.append('tag', tag as string);
  params.append('limit', '6');

  const url = `${process.env.NEXT_PUBLIC_BACKEND_URL}/blog/posts?${params.toString()}`;
  console.log(`Fetching posts from URL: ${url}`); // Debugging

  const res = await fetch(url);
  if (!res.ok) throw new Error('Network response was not ok');

  const data: ApiResponse = await res.json();
  console.log('Received data:', data); // Debugging

  return data;
};

const InfinitePosts: React.FC = () => {
  const searchParams = useSearchParams();
  const currentTag = searchParams.get('tag') || '';

  const {
    data,
    error,
    fetchNextPage,
    hasNextPage,
    isFetching,
    isPending,
    status,
  } = useInfiniteQuery<ApiResponse, Error, InfiniteData<ApiResponse>, PostsQueryKey, string | undefined>({
    queryKey: ['posts', currentTag] as PostsQueryKey,
    queryFn: fetchPosts,
    getNextPageParam: (lastPage) => lastPage.nextCursor || undefined,
    initialPageParam: undefined, // Explicitly set initialPageParam
    staleTime: 1000 * 60 * 5, // 5 minutes
  });

  // Flatten all posts from pages
  const allPosts = data?.pages.flatMap(page => page.items || []) || [];

  return (
    <>
      {isPending && <p>Loading...</p>}
      {status === 'error' && <p>failed to load</p>}

      {/* Handle Success State with InfiniteScroll */}
      {status === 'success' && (
        <InfiniteScroll
          dataLength={allPosts.length}
          next={() => fetchNextPage()}
          hasMore={!!hasNextPage}
          loader={<p>Loading more posts...</p>}
          endMessage={<p>No more posts to load.</p>}
          scrollThreshold={0.9} // Trigger fetch when 90% scrolled
        >
          <ul className="flex flex-col">
            {allPosts.map((post: Post, index: number) => (
              <li key={`${post.slug}-${index}`}> {/* Ensure unique keys */}
                <PostItem
                  slug={post.slug}
                  date={post.created_at}
                  title={post.title}
                  description={post.description}
                  tags={post.tags}
                />
              </li>
            ))}
          </ul>
        </InfiniteScroll>
      )}
    </>
  );
};

export default InfinitePosts;
