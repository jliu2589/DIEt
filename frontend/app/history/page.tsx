"use client";

import { useEffect, useState } from "react";
import { getRecentMeals, type RecentMeal } from "@/lib/api";

const USER_ID = "demo-user";

export default function HistoryPage() {
  const [meals, setMeals] = useState<RecentMeal[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [endpointMissing, setEndpointMissing] = useState(false);

  useEffect(() => {
    async function load() {
      setLoading(true);
      setError(null);
      try {
        const data = await getRecentMeals(USER_ID, 20);
        setMeals(data);
      } catch (e) {
        const msg = e instanceof Error ? e.message : "Failed to load recent meals";
        if (msg === "RECENT_MEALS_ENDPOINT_NOT_IMPLEMENTED") {
          setEndpointMissing(true);
        } else {
          setError(msg);
        }
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

      {endpointMissing && (
        <section className="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900">
          Recent meals endpoint is not yet available on the backend.
          <br />
          Expected endpoint: <code>GET /v1/meals/recent?user_id=&lt;id&gt;&amp;limit=20</code>
        </section>
      )}

      {!endpointMissing && (
        <section className="overflow-hidden rounded-xl bg-white shadow-sm">
          <table className="min-w-full text-sm">
            <thead className="bg-slate-100 text-left text-slate-700">
              <tr>
                <th className="px-4 py-3 font-medium">Meal</th>
                <th className="px-4 py-3 font-medium">Timestamp</th>
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
                  <td className="px-4 py-4 text-slate-500" colSpan={6}>
                    No meals found.
                  </td>
                </tr>
              )}

              {!loading &&
                meals.map((meal) => (
                  <tr key={meal.id} className="border-t border-slate-100">
                    <td className="px-4 py-3">{meal.canonical_name}</td>
                    <td className="px-4 py-3">{new Date(meal.eaten_at).toLocaleString()}</td>
                    <td className="px-4 py-3">{meal.calories_kcal}</td>
                    <td className="px-4 py-3">{meal.protein_g} g</td>
                    <td className="px-4 py-3">{meal.carbohydrate_g} g</td>
                    <td className="px-4 py-3">{meal.fat_g} g</td>
                  </tr>
                ))}
            </tbody>
          </table>
        </section>
      )}

      {error && <p className="text-sm text-red-600">{error}</p>}
    </main>
  );
}
