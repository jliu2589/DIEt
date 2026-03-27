"use client";

import { useState } from "react";
import { getDailySummary, type DailySummaryResponse } from "@/lib/api";

function isoDate(date: Date) {
  return date.toISOString().slice(0, 10);
}

export default function HistoryPage() {
  const [userId, setUserId] = useState("demo-user");
  const [date, setDate] = useState(isoDate(new Date()));
  const [result, setResult] = useState<DailySummaryResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onLoad() {
    setLoading(true);
    setError(null);
    try {
      const summary = await getDailySummary(userId, date);
      setResult(summary);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load summary");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="space-y-6">
      <section className="rounded-xl bg-white p-4 shadow-sm">
        <h2 className="mb-3 text-base font-semibold">Meal History (Daily)</h2>
        <p className="mb-4 text-sm text-slate-600">
          V1 shows daily totals by date. Detailed meal event history can be added later.
        </p>
        <div className="flex flex-wrap items-end gap-3">
          <label className="text-sm">
            <span className="mb-1 block text-slate-600">User ID</span>
            <input
              value={userId}
              onChange={(e) => setUserId(e.target.value)}
              className="rounded-md border border-slate-300 px-3 py-2"
            />
          </label>
          <label className="text-sm">
            <span className="mb-1 block text-slate-600">Date</span>
            <input
              type="date"
              value={date}
              onChange={(e) => setDate(e.target.value)}
              className="rounded-md border border-slate-300 px-3 py-2"
            />
          </label>
          <button onClick={onLoad} className="rounded-md bg-slate-900 px-4 py-2 text-sm text-white">
            {loading ? "Loading..." : "Load"}
          </button>
        </div>
      </section>

      {result && (
        <section className="rounded-xl bg-white p-4 shadow-sm">
          <h3 className="mb-3 text-base font-semibold">{result.date} Totals</h3>
          <ul className="space-y-1 text-sm">
            <li>Calories: {result.calories_kcal} kcal</li>
            <li>Protein: {result.protein_g} g</li>
            <li>Carbohydrate: {result.carbohydrate_g} g</li>
            <li>Fat: {result.fat_g} g</li>
          </ul>
        </section>
      )}

      {error && <p className="text-sm text-red-600">{error}</p>}
    </main>
  );
}
