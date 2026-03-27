import "./globals.css";
import Link from "next/link";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Nutrition Tracker",
  description: "Minimal nutrition tracking dashboard"
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <div className="mx-auto min-h-screen w-full max-w-4xl px-4 py-6">
          <header className="mb-6 flex items-center justify-between rounded-xl bg-white px-4 py-3 shadow-sm">
            <h1 className="text-lg font-semibold">Nutrition Tracker</h1>
            <nav className="flex gap-4 text-sm">
              <Link href="/" className="hover:text-blue-600">
                Today
              </Link>
              <Link href="/history" className="hover:text-blue-600">
                History
              </Link>
            </nav>
          </header>
          {children}
        </div>
      </body>
    </html>
  );
}
