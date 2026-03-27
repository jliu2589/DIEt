"use client";

import { FormEvent, useMemo, useState } from "react";
import { createMeal, getDailySummary, type CreateMealResponse, type DailySummaryResponse } from "@/lib/api";

function todayDate() {
  return new Date().toISOString().slice(0, 10);
}

export default function HomePage() {
  const [userId, setUserId] = useState("demo-user");
  const [rawText, setRawText] = useState("");
  const [summary, setSummary] = useState<DailySummaryResponse | null>(null);
  const [lastMeal, setLastMeal] = useState<CreateMealResponse | null>(null);
  const [loadingSummary, setLoadingSummary] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const date = useMemo(() => todayDate(), []);

  async function loadSummary() {
    setLoadingSummary(true);
    setError(null);
    try {
      const data = await getDailySummary(userId, date);
      setSummary(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to fetch summary");
    } finally {
      setLoadingSummary(false);
    }
  }

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!rawText.trim()) return;

    setSubmitting(true);
    setError(null);
    try {
      const result = await createMeal({
        user_id: userId,
        source: "web",
        raw_text: rawText,
        eaten_at: new Date().toISOString()
      });
      setLastMeal(result);
      setRawText("");
      await loadSummary();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to add meal");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="space-y-6">
      <section className="rounded-xl bg-white p-4 shadow-sm">
        <h2 className="mb-3 text-base font-semibold">Today Dashboard</h2>
        <div className="flex flex-wrap items-end gap-3">
          <label className="text-sm">
            <span className="mb-1 block text-slate-600">User ID</span>
            <input
              value={userId}
              onChange={(e) => setUserId(e.target.value)}
              className="rounded-md border border-slate-300 px-3 py-2"
            />
          </label>
          <button onClick={loadSummary} className="rounded-md bg-slate-900 px-4 py-2 text-sm text-white">
            {loadingSummary ? "Loading..." : "Load Summary"}
          </button>
        </div>
      </section>

      <section className="rounded-xl bg-white p-4 shadow-sm">
        <h3 className="mb-3 text-base font-semibold">Log Meal</h3>
        <form onSubmit={onSubmit} className="space-y-3">
          <textarea
            value={rawText}
            onChange={(e) => setRawText(e.target.value)}
            placeholder="e.g. grilled salmon with rice and broccoli"
            className="min-h-24 w-full rounded-md border border-slate-300 px-3 py-2"
          />
          <button
            type="submit"
            disabled={submitting}
            className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white disabled:opacity-60"
          >
            {submitting ? "Saving..." : "Add Meal"}
          </button>
        </form>
      </section>

      {lastMeal && (
        <section className="rounded-xl bg-white p-4 shadow-sm">
          <h3 className="mb-2 text-base font-semibold">Last Processed Meal</h3>
          <p className="text-sm text-slate-700">
            <strong>{lastMeal.canonical_name}</strong> via {lastMeal.processed_from}
          </p>
        </section>
      )}

      {summary && (
        <section className="rounded-xl bg-white p-4 shadow-sm">
          <h3 className="mb-3 text-base font-semibold">Totals ({summary.date})</h3>
          <div className="grid grid-cols-2 gap-3 text-sm md:grid-cols-4">
            <Metric label="Calories" value={`${summary.calories_kcal} kcal`} />
            <Metric label="Protein" value={`${summary.protein_g} g`} />
            <Metric label="Carbs" value={`${summary.carbohydrate_g} g`} />
            <Metric label="Fat" value={`${summary.fat_g} g`} />
          </div>
        </section>
      )}

      {error && <p className="text-sm text-red-600">{error}</p>}
    </main>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-slate-200 p-3">
      <p className="text-slate-500">{label}</p>
      <p className="mt-1 font-medium">{value}</p>
    </div>
  );
}
