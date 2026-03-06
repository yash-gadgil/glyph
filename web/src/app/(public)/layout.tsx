import PixelEtherBackground from "@/components/background/PixelEtherBackground";

export default function PublicLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {


  return (
    <div>
      <PixelEtherBackground />
      {children}
    </div>
  );
}