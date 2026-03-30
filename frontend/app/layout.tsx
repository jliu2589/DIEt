import "./globals.css";
import Link from "next/link";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Nutrition Tracker",
  description: "Warm, minimal nutrition dashboard"
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <div className="mx-auto min-h-screen w-full max-w-5xl px-4 py-6 sm:px-6 lg:py-8">
          <header className="mb-6 rounded-2xl border border-stone-200 bg-stone-50/80 px-4 py-3 shadow-sm">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <p className="text-xs uppercase tracking-[0.14em] text-stone-500">Nutrition App</p>
                <h1 className="text-xl font-semibold tracking-tight text-stone-900">Daily Dashboard</h1>
              </div>
              <nav className="flex gap-3 text-sm text-stone-600">
                <Link href="/" className="rounded-lg px-2 py-1 hover:bg-stone-200/60 hover:text-stone-900">
                  Today
                </Link>
                <Link href="/history" className="rounded-lg px-2 py-1 hover:bg-stone-200/60 hover:text-stone-900">
                  History
                </Link>
              </nav>
            </div>
          </header>
          {children}
        </div>
      </body>
    </html>
  );
}
