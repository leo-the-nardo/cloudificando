// /components/tag.tsx
'use client'
import Link from "next/link";
import { slug } from "github-slugger";
import { badgeVariants } from "./ui/badge";
import { useSearchParams } from "next/navigation";
import { useTags } from "@/hooks/use-tags";
import { Suspense } from "react";

interface TagProps {
  tag: string;
  current?: boolean;
  count?: number;
}

export function TagList () {
  const currentTag = useSearchParams().get('tag') || '';
  const { data:tags ,isError,isPending } = useTags()

  if (isPending) {
    return <div>Loading...</div>;
  }
  if (isError) {
    return <div></div>;
  }
  return (
    <>
        <Suspense>
      {tags && tags.map((tag) => (
        <Tag
          tag={tag.tag}
          key={tag.tag}
          count={tag.count}
          current={currentTag === tag.tag}
        />
      ))}
      </Suspense>
    </>
  );
}
export function Tag({ tag, current, count }: TagProps) {
  return (
    <Link
      className={badgeVariants({
        variant: current ? "default" : "secondary",
        className: "no-underline rounded-md",
      })}
      href={`/posts?tag=${slug(tag)}`}
      aria-current={current ? "page" : undefined} // accessibility
    >
      {tag} {count ? `(${count})` : null}
    </Link>
  );
}
