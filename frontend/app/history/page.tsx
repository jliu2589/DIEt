"use client";

import { useEffect, useState } from "react";
import { getRecentMeals, type RecentMeal } from "@/lib/api";

const USER_ID = "demo-user";

function formatDateTime(value: string) {
  return new Date(value).toLocaleString();
}

function renderMacro(value: number | null) {
  return value == null ? "—" : `${value} g`;
}

export default function HistoryPage() {
  const [meals, setMeals] = useState<RecentMeal[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      setLoading(true);
      setError(null);
      try {
        const data = await getRecentMeals(USER_ID, 20);
        setMeals(data.items ?? []);
      } catch {
        setMeals([]);
        setError("We couldn't load your recent meals right now. Please try again shortly.");
      } finally {
        setLoading(false);
      }
    }

    void load();
  }, []);

  return (
    <main className="space-y-4">
      <section className="rounded-xl bg-white p-4 shadow-sm">
        <h2 className="text-base font-semibold">Meal History</h2>
        <p className="mt-1 text-sm text-slate-600">Recent meals for user: {USER_ID}</p>
      </section>

      <section className="overflow-hidden rounded-xl bg-white shadow-sm">
        <table className="min-w-full text-sm">
          <thead className="bg-slate-100 text-left text-slate-700">
            <tr>
              <th className="px-4 py-3 font-medium">Meal</th>
              <th className="px-4 py-3 font-medium">Time</th>
              <th className="px-4 py-3 font-medium">Calories</th>
              <th className="px-4 py-3 font-medium">Protein</th>
              <th className="px-4 py-3 font-medium">Carbs</th>
              <th className="px-4 py-3 font-medium">Fat</th>
            </tr>
          </thead>
          <tbody>
            {loading && (
              <tr>
                <td className="px-4 py-4 text-slate-500" colSpan={6}>
                  Loading...
                </td>
              </tr>
            )}

            {!loading && meals.length === 0 && (
              <tr>
                <td className="px-4 py-5 text-center text-slate-500" colSpan={6}>
                  No meals yet. Start by logging your first meal.
                </td>
              </tr>
            )}

            {!loading &&
              meals.map((meal) => (
                <tr key={meal.meal_event_id} className="border-t border-slate-100">
                  <td className="px-4 py-3">{meal.canonical_name}</td>
                  <td className="px-4 py-3">
                    <div>{formatDateTime(meal.eaten_at)}</div>
                    <div className="text-xs text-slate-500">Logged: {formatDateTime(meal.logged_at)}</div>
                    {meal.time_source && <div className="text-xs text-slate-500">Source: {meal.time_source}</div>}
                  </td>
                  <td className="px-4 py-3">{meal.calories_kcal ?? "—"}</td>
                  <td className="px-4 py-3">{renderMacro(meal.protein_g)}</td>
                  <td className="px-4 py-3">{renderMacro(meal.carbohydrate_g)}</td>
                  <td className="px-4 py-3">{renderMacro(meal.fat_g)}</td>
                </tr>
              ))}
          </tbody>
        </table>
      </section>

      {error && (
        <p className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900">{error}</p>
      )}
    </main>
  );
}
