import Navbar from "@/components/ui/Navbar";
import ChatWidget from "@/components/advisor/ChatWidget";

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {


  return (
    <div className="font-mono bg-background min-h-screen">
      <Navbar />

      {children}

      <ChatWidget />
    </div>
  );
}