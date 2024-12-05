import { buttonVariants } from "@/components/ui/button";
import { siteConfig } from "@/config/site";
import { cn } from "@/lib/utils";
import Link from "next/link";
import { PostItem } from "@/components/post-item";
import Image from "next/image";
import { fetchPosts } from "@/lib/posts";

export default async function Home() {
  const latestPosts = await fetchPosts();
  return (
    <>
      <section className=" relative space-y-6 pb-8 pt-6 md:pb-12  lg:py-24 lg:pb-14">
          <div className=" justify-center container flex items-center ">
            <Image src="/cloudinho.png" alt="Cloudificando" className="hidden md:block" width={340} height={340}/>
            {/* <Image src="/seta.png" alt="Cloudificando"     className="absolute top-14 lg:top-1/4 left-300  md:block" width={240} height={240}/> */}
            <div className="flex items-start flex-col gap-4">
              <h1 className="text-3xl sm:text-5xl md:text-6xl lg:text-7xl font-black text-balance">
              Cloudificando
              </h1>
              <p className="max-w-[42rem] text-muted-foreground sm:text-xl text-balance">
              It&apos;s not beautiful, but it works, just like me
              </p>
            <div className="flex flex-col gap-4 justify-center sm:flex-row">
            <Link
              href="/posts"
              className={cn(buttonVariants({ size: "lg" }), "w-full sm:w-fit")}
            >
              View my blog
            </Link>
            <Link
              href={siteConfig.links.github}
              target="_blank"
              rel="noreferrer"
              className={cn(
                buttonVariants({ variant: "outline", size: "lg" }),
                "w-full sm:w-fit"
              )}
            >
              GitHub
            </Link>
            </div>
          </div>
        </div>
      </section>
      <section className="container max-w-4xl py-6 lg:py-10 flex flex-col space-y-6 mt-60">
        <h2 className="text-3xl sm:text-5xl md:text-6xl lg:text-7xl font-black text-center">
          Latest Posts
        </h2>
        <ul className="flex flex-col">
          {latestPosts.items.map((post) => (
              <li key={post.slug} className="first:border-t first:border-border">
                <PostItem
                  slug={post.slug}
                  title={post.title}
                  description={post.description!}
                  date={post.created_at}
                  tags={post.tags}
                />
              </li>
          ))}
        </ul>
      </section>
    </>
  );
}
