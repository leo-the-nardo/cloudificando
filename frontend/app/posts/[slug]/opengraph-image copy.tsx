import { ImageResponse } from 'next/og'
  
// Image metadata
export const alt = 'About Acme'
export const size = {
  width: 1200,
  height: 630,
}

export const contentType = 'image/png'
 
// Image generation
export default async function Image({ params }: { params: { slug: string } }) {
  // Font
  // const interSemiBold = await (await fetch(`https://rsms.me/inter/font-files/Inter-Bold.woff`)).arrayBuffer()
  // const interSemiBold = fetch(
  //   new URL('../../../assets/fonts/Inter-Bold.ttf', import.meta.url)
  // ).then((res) => res.arrayBuffer())
 
  return new ImageResponse(
    (
      // ImageResponse JSX element
      <div
        style={{
          fontSize: 48,
          background: 'white',
          width: '100%',
          height: '100%',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          // font
        }}
      >
        About Acme - {params.slug}
      </div>
    ),
    // ImageResponse options
    {
      // For convenience, we can re-use the exported opengraph-image
      // size config to also set the ImageResponse's width and height.
      ...size,
      // fonts: await getFonts(),
      // fonts: [
      //   {
      //     name: 'Inter',
      //     data: await interSemiBold,
      //     style: 'normal',
      //     weight: 400,
      //   },
      // ],
    }
  )
}

// import type { Font } from "satori";

async function getFonts(): Promise<any> {
  // This is unfortunate but I can't figure out how to load local font files
  // when deployed to vercel.
  const [interRegular, interMedium, interSemiBold, interBold] =
    await Promise.all([
      fetch(`https://rsms.me/inter/font-files/Inter-Regular.woff`).then((res) =>
        res.arrayBuffer()
      ),
      fetch(`https://rsms.me/inter/font-files/Inter-Medium.woff`).then((res) =>
        res.arrayBuffer()
      ),
      fetch(`https://rsms.me/inter/font-files/Inter-SemiBold.woff`).then(
        (res) => res.arrayBuffer()
      ),
      fetch(`https://rsms.me/inter/font-files/Inter-Bold.woff`).then((res) =>
        res.arrayBuffer()
      ),
    ]);

  return [
    {
      name: "Inter",
      data: interRegular,
      style: "normal",
      weight: 400,
    },
    {
      name: "Inter",
      data: interMedium,
      style: "normal",
      weight: 500,
    },
    {
      name: "Inter",
      data: interSemiBold,
      style: "normal",
      weight: 600,
    },
    {
      name: "Inter",
      data: interBold,
      style: "normal",
      weight: 700,
    },
  ];
}
