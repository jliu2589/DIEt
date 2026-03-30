"use client";

import { useEffect, useState } from "react";
import { editMealTime, getRecentMeals, type RecentMeal } from "@/lib/api";

const USER_ID = "demo-user";

function formatPrimaryMealTime(value: string) {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return "—";
  }
  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit"
  }).format(parsed);
}

function getTimeHint(timeSource?: string | null) {
  if (!timeSource) {
    return null;
  }
  if (timeSource === "default_now") {
    return "Logged for now. Edit time if needed.";
  }
  return null;
}

function renderMacro(value: number | null) {
  return value == null ? "—" : `${value} g`;
}

export default function HistoryPage() {
  const [meals, setMeals] = useState<RecentMeal[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editingMealID, setEditingMealID] = useState<number | null>(null);
  const [editingValue, setEditingValue] = useState("");
  const [savingMealID, setSavingMealID] = useState<number | null>(null);
  const [editError, setEditError] = useState<string | null>(null);

  async function loadMeals() {
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

  useEffect(() => {
    void loadMeals();
  }, []);

  function toDateTimeLocalValue(timestamp: string) {
    const parsed = new Date(timestamp);
    if (Number.isNaN(parsed.getTime())) {
      return "";
    }
    const pad = (value: number) => String(value).padStart(2, "0");
    return `${parsed.getFullYear()}-${pad(parsed.getMonth() + 1)}-${pad(parsed.getDate())}T${pad(parsed.getHours())}:${pad(parsed.getMinutes())}`;
  }

  async function onSaveMealTime(mealEventID: number) {
    if (!editingValue) {
      return;
    }
    setSavingMealID(mealEventID);
    setEditError(null);
    try {
      await editMealTime(mealEventID, {
        user_id: USER_ID,
        eaten_at: new Date(editingValue).toISOString()
      });
      setEditingMealID(null);
      setEditingValue("");
      await loadMeals();
    } catch {
      setEditError("Could not update meal time right now. Please try again.");
    } finally {
      setSavingMealID(null);
    }
  }

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
              meals.map((meal) => {
                const timeHint = getTimeHint(meal.time_source);
                return (
                  <tr key={meal.meal_event_id} className="border-t border-slate-100">
                    <td className="px-4 py-3">{meal.canonical_name}</td>
                    <td className="px-4 py-3">
                      <div>{formatPrimaryMealTime(meal.eaten_at)}</div>
                      {timeHint && <div className="text-xs text-slate-500">{timeHint}</div>}
                      {editingMealID === meal.meal_event_id ? (
                        <div className="mt-2 flex items-center gap-2">
                          <input
                            type="datetime-local"
                            value={editingValue}
                            onChange={(event) => setEditingValue(event.target.value)}
                            className="rounded-md border border-slate-300 px-2 py-1 text-xs text-slate-700"
                          />
                          <button
                            type="button"
                            onClick={() => void onSaveMealTime(meal.meal_event_id)}
                            disabled={savingMealID === meal.meal_event_id || !editingValue}
                            className="rounded-md bg-slate-900 px-2 py-1 text-xs font-medium text-white"
                          >
                            {savingMealID === meal.meal_event_id ? "Saving..." : "Save"}
                          </button>
                          <button
                            type="button"
                            onClick={() => {
                              setEditingMealID(null);
                              setEditingValue("");
                            }}
                            className="text-xs text-slate-600"
                          >
                            Cancel
                          </button>
                        </div>
                      ) : (
                        <button
                          type="button"
                          onClick={() => {
                            setEditingMealID(meal.meal_event_id);
                            setEditingValue(toDateTimeLocalValue(meal.eaten_at));
                            setEditError(null);
                          }}
                          className="mt-1 text-xs font-medium text-slate-600 underline decoration-slate-300 underline-offset-2"
                        >
                          Edit time
                        </button>
                      )}
                    </td>
                    <td className="px-4 py-3">{meal.calories_kcal ?? "—"}</td>
                    <td className="px-4 py-3">{renderMacro(meal.protein_g)}</td>
                    <td className="px-4 py-3">{renderMacro(meal.carbohydrate_g)}</td>
                    <td className="px-4 py-3">{renderMacro(meal.fat_g)}</td>
                  </tr>
                );
              })}
          </tbody>
        </table>
      </section>

      {error && (
        <p className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900">{error}</p>
      )}
      {editError && <p className="rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-800">{editError}</p>}
    </main>
  );
}
