/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: false, // Set this to false
  
  // headers: async () => {
  //   return [
  //     {
  //       source: "/posts",
  //       headers: [
  //         {
  //           key: "Cache-Control",
  //           value: "s-maxage=86400, max-age=0, must-revalidate",
  //         },
  //       ],
  //     },
  //   ];
  // },
};


export default nextConfig;
