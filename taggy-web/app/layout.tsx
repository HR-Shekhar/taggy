import type { Metadata } from "next";
import { Antic, JetBrains_Mono } from "next/font/google";
import { AuthProvider } from "@/lib/auth";
import { ThemeProvider } from "@/components/theme-provider";
import { PageBackground } from "@/components/page-background";
import { Toaster } from "@/components/ui/sonner";
import "@/styles/globals.css";
import { cn } from "@/lib/utils";

const antic = Antic({
  weight: "400",
  subsets: ["latin"],
  variable: "--font-sans",
});

const jetbrains = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
});

export const metadata: Metadata = {
  title: "Taggy: Gym buddies for skill growth",
  description:
    "Structured roadmaps, accountability pods, community chat, and progress tracking.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html
      lang="en"
      className={cn(antic.variable, jetbrains.variable)}
      suppressHydrationWarning
    >
      <body className="min-h-screen bg-transparent font-sans antialiased">
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <PageBackground />
          <div className="relative z-10 min-h-dvh">
            <AuthProvider>{children}</AuthProvider>
          </div>
          <Toaster richColors closeButton position="top-center" />
        </ThemeProvider>
      </body>
    </html>
  );
}
