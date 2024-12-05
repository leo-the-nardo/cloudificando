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
  const {slug} = params
  const GITHUB_REPO = process.env.CONTENT_GITHUB_REPO
  const GITHUB_BRANCH = process.env.CONTENT_GITHUB_BRANCH
  const GITHUB_MDX_PATH = process.env.CONTENT_GITHUB_MDX_PATH
  const imgUrl = `https://raw.githubusercontent.com/${GITHUB_REPO}/refs/heads/${GITHUB_BRANCH}/${GITHUB_MDX_PATH}/open-graph/${slug}.png`;
  return new ImageResponse(
    (
      <img src={imgUrl}/>
    ),
    {
      ...size,
    }
  )
}

