import PixelEtherBackground from "@/components/background/PixelEtherBackground";

export default async function AppLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="bg-black">
      <PixelEtherBackground />
      {children}
    </div>
  );
}