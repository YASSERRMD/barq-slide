import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "barq-slides — AI Presentation Engine",
  description:
    "Next-gen AI presentation engine with dual-render architecture: real-time HTML preview and native PPTX export.",
};

import { Navigation } from "@/components/navigation";

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={inter.className}>
        <Navigation />
        <div className="pt-14">
          {children}
        </div>
      </body>
    </html>
  );
}
