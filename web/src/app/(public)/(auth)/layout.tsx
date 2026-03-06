import GlassSurface from "@/components/primitives/GlassSurface";
import BackButton from "@/components/ui/BackButton";

export default function AuthLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="h-screen flex justify-center items-center pointer-events-none font-mono">
      <GlassSurface
        displace={15}
        distortionScale={-150}
        redOffset={5}
        greenOffset={15}
        blueOffset={25}
        brightness={10}
        opacity={0.1}
        mixBlendMode="overlay"
        className="w-6/7 h-auto py-18 max-w-150"
      >
        <BackButton />
        {children}
      </GlassSurface>
    </div>
  );
}