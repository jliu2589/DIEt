"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { createMeal, getDailySummary, type CreateMealResponse, type DailySummaryResponse } from "@/lib/api";

const USER_ID = "demo-user";

function todayDate() {
  return new Date().toISOString().slice(0, 10);
}

export default function HomePage() {
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
      const data = await getDailySummary(USER_ID, date);
      setSummary(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to fetch summary");
    } finally {
      setLoadingSummary(false);
    }
  }

  useEffect(() => {
    void loadSummary();
  }, []);

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!rawText.trim()) return;

    setSubmitting(true);
    setError(null);
    try {
      const result = await createMeal({
        user_id: USER_ID,
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
        <h2 className="mb-2 text-base font-semibold">Today Dashboard</h2>
        <p className="mb-4 text-sm text-slate-600">User: {USER_ID}</p>

        <form onSubmit={onSubmit} className="flex gap-2">
          <input
            value={rawText}
            onChange={(e) => setRawText(e.target.value)}
            placeholder="e.g. grilled salmon with rice"
            className="w-full rounded-md border border-slate-300 px-3 py-2"
          />
          <button
            type="submit"
            disabled={submitting}
            className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white disabled:opacity-60"
          >
            {submitting ? "Saving..." : "Log Meal"}
          </button>
        </form>
      </section>

      {lastMeal && (
        <section className="rounded-xl border border-emerald-200 bg-emerald-50 p-4 shadow-sm">
          <h3 className="mb-2 text-base font-semibold text-emerald-900">Meal Logged</h3>
          <p className="mb-2 text-sm text-emerald-900">
            <strong>{lastMeal.canonical_name}</strong>
          </p>
          <div className="grid grid-cols-2 gap-2 text-sm md:grid-cols-4">
            <Metric label="Calories" value={`${lastMeal.nutrition.calories_kcal} kcal`} />
            <Metric label="Protein" value={`${lastMeal.nutrition.protein_g} g`} />
            <Metric label="Carbs" value={`${lastMeal.nutrition.carbohydrate_g} g`} />
            <Metric label="Fat" value={`${lastMeal.nutrition.fat_g} g`} />
          </div>
        </section>
      )}

      <section className="rounded-xl bg-white p-4 shadow-sm">
        <div className="mb-3 flex items-center justify-between">
          <h3 className="text-base font-semibold">Today's Summary ({date})</h3>
          <button onClick={loadSummary} className="text-sm text-blue-600 hover:underline">
            {loadingSummary ? "Refreshing..." : "Refresh"}
          </button>
        </div>

        {summary ? (
          <div className="grid grid-cols-2 gap-3 text-sm md:grid-cols-3">
            <Metric label="Calories" value={`${summary.calories_kcal} kcal`} />
            <Metric label="Protein" value={`${summary.protein_g} g`} />
            <Metric label="Carbs" value={`${summary.carbohydrate_g} g`} />
            <Metric label="Fat" value={`${summary.fat_g} g`} />
            <Metric label="Fiber" value={`${summary.fiber_g} g`} />
            <Metric label="Sodium" value={`${summary.sodium_mg} mg`} />
            <Metric label="Potassium" value={`${summary.potassium_mg} mg`} />
            <Metric label="Calcium" value={`${summary.calcium_mg} mg`} />
            <Metric label="Magnesium" value={`${summary.magnesium_mg} mg`} />
          </div>
        ) : (
          <p className="text-sm text-slate-500">No summary data yet.</p>
        )}
      </section>

      {error && <p className="text-sm text-red-600">{error}</p>}
    </main>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-slate-200 bg-white p-3">
      <p className="text-slate-500">{label}</p>
      <p className="mt-1 font-medium">{value}</p>
    </div>
  );
}
