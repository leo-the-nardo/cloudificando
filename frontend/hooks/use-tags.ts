import { useQuery } from "@tanstack/react-query";

export type FetchTagsResponse = {
  tag: string;
  count: number;
}[]


async function fetchTags(): Promise<FetchTagsResponse> {
  const response = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/blog/tags`, {
    next: { tags: ['tags']},
  });
  if (!response.ok) {
    throw new Error('Failed to fetch tags');
  }
  const data = await response.json();
  return data;
}

export function useTags() {
  return useQuery<FetchTagsResponse>({queryKey: ["tags"], queryFn: () => fetchTags()});
}