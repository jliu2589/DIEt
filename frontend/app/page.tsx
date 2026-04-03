"use client";

import { FormEvent, useEffect, useState } from "react";
import { deleteMeal, editMealTime, getDashboardToday, getRecentMeals, getTrends, postChat, type ChatResponse, type RecentMeal, type TrendsPoint } from "../lib/api";

const DASHBOARD_USER_ID = "demo-user-001";

const goalTargets = [
  { label: "Calories", key: "calories" as const, target: 2500, unit: "kcal" },
  { label: "Protein", key: "protein" as const, target: 160, unit: "g" },
  { label: "Carbs", key: "carbs" as const, target: 220, unit: "g" },
  { label: "Fat", key: "fat" as const, target: 75, unit: "g" }
];

type TrendMetric = "calories" | "protein" | "carbs" | "fat" | "weight";
type TrendRange = "weekly" | "monthly" | "90d" | "yearly";

const metricLabels: Record<TrendMetric, string> = {
  calories: "Calories",
  protein: "Protein",
  carbs: "Carbs",
  fat: "Fat",
  weight: "Weight"
};

const rangeLabels: Record<TrendRange, string> = {
  weekly: "Weekly",
  monthly: "Monthly",
  "90d": "90 days",
  yearly: "Yearly"
};

const rangeToApi: Record<TrendRange, "7d" | "30d" | "90d" | "1y"> = {
  weekly: "7d",
  monthly: "30d",
  "90d": "90d",
  yearly: "1y"
};

export default function HomePage() {
  const [goals, setGoals] = useState(
    goalTargets.map((goal) => ({
      label: goal.label,
      consumed: 0,
      target: goal.target,
      unit: goal.unit
    }))
  );
  const [chatInput, setChatInput] = useState("");
  const [chatTurns, setChatTurns] = useState<Array<{ id: number; userMessage: string; response: ChatResponse }>>([]);
  const [recentMeals, setRecentMeals] = useState<RecentMeal[]>([]);
  const [dashboardDate, setDashboardDate] = useState<string>(currentLocalISODate());
  const [mealsError, setMealsError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [chatError, setChatError] = useState<string | null>(null);
  const [dashboardError, setDashboardError] = useState<string | null>(null);
  const [isDashboardLoading, setIsDashboardLoading] = useState(true);
  const [activeMetric, setActiveMetric] = useState<TrendMetric>("calories");
  const [activeRange, setActiveRange] = useState<TrendRange>("weekly");
  const [trendPointsByRange, setTrendPointsByRange] = useState<Partial<Record<TrendRange, TrendsPoint[]>>>({});
  const [trendsError, setTrendsError] = useState<string | null>(null);
  const [isTrendsLoading, setIsTrendsLoading] = useState(false);
  const [editingMealID, setEditingMealID] = useState<number | null>(null);
  const [editingMealValue, setEditingMealValue] = useState("");
  const [savingMealID, setSavingMealID] = useState<number | null>(null);
  const [deletingMealID, setDeletingMealID] = useState<number | null>(null);
  const [editMealError, setEditMealError] = useState<string | null>(null);

  async function refreshDashboard() {
    setDashboardError(null);
    setMealsError(null);
    try {
      const dashboard = await getDashboardToday(DASHBOARD_USER_ID);
      const localToday = currentLocalISODate();
      setDashboardDate(localToday);
      const totals = dashboard.daily_summary.totals;
      setGoals([
        { label: "Calories", consumed: Math.round(totals.calories_kcal), target: 2500, unit: "kcal" },
        { label: "Protein", consumed: Math.round(totals.protein_g), target: 160, unit: "g" },
        { label: "Carbs", consumed: Math.round(totals.carbohydrate_g), target: 220, unit: "g" },
        { label: "Fat", consumed: Math.round(totals.fat_g), target: 75, unit: "g" }
      ]);
      if (Array.isArray(dashboard.recent_meals)) {
        setRecentMeals(dashboard.recent_meals.filter((meal) => isMealOnDate(meal.eaten_at, localToday)));
      } else {
        const recent = await getRecentMeals(DASHBOARD_USER_ID, 20);
        setRecentMeals(recent.items.filter((meal) => isMealOnDate(meal.eaten_at, localToday)));
      }
    } catch (error) {
      setDashboardError(error instanceof Error ? error.message : "Failed to load dashboard");
      try {
        const recent = await getRecentMeals(DASHBOARD_USER_ID, 20);
        const todayIso = currentLocalISODate();
        setDashboardDate(todayIso);
        setRecentMeals(recent.items.filter((meal) => isMealOnDate(meal.eaten_at, todayIso)));
      } catch (mealsFetchError) {
        setMealsError(mealsFetchError instanceof Error ? mealsFetchError.message : "Failed to load recent meals");
      }
    } finally {
      setIsDashboardLoading(false);
    }
  }

  useEffect(() => {
    void refreshDashboard();
  }, []);

  useEffect(() => {
    const interval = window.setInterval(() => {
      const todayIso = currentLocalISODate();
      if (todayIso !== dashboardDate) {
        setTrendPointsByRange({});
        void refreshDashboard();
      }
    }, 60_000);

    return () => {
      window.clearInterval(interval);
    };
  }, [dashboardDate]);

  useEffect(() => {
    const alreadyLoaded = trendPointsByRange[activeRange];
    if (alreadyLoaded) {
      return;
    }
    setIsTrendsLoading(true);
    setTrendsError(null);
    void getTrends(DASHBOARD_USER_ID, rangeToApi[activeRange])
      .then((response) => {
        setTrendPointsByRange((prev) => ({ ...prev, [activeRange]: response.points }));
      })
      .catch((error: unknown) => {
        setTrendsError(error instanceof Error ? error.message : "Failed to load trends");
      })
      .finally(() => {
        setIsTrendsLoading(false);
      });
  }, [activeRange, trendPointsByRange]);

  async function onSubmitChat(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmed = chatInput.trim();
    if (!trimmed) {
      return;
    }
    setIsSubmitting(true);
    setChatError(null);

    try {
      const response = await postChat(DASHBOARD_USER_ID, trimmed);
      setChatTurns((prev) => [{ id: Date.now(), userMessage: trimmed, response }, ...prev].slice(0, 6));
      setChatInput("");
      await refreshDashboard();
    } catch (error) {
      setChatError(error instanceof Error ? error.message : "Could not send chat message");
    } finally {
      setIsSubmitting(false);
    }
  }

  async function onSaveMealTime(mealEventID: number) {
    if (!editingMealValue) {
      return;
    }
    setSavingMealID(mealEventID);
    setEditMealError(null);
    try {
      await editMealTime(mealEventID, {
        user_id: DASHBOARD_USER_ID,
        eaten_at: new Date(editingMealValue).toISOString()
      });
      setEditingMealID(null);
      setEditingMealValue("");
      await refreshDashboard();
    } catch (error) {
      setEditMealError(error instanceof Error ? error.message : "Could not update meal time");
    } finally {
      setSavingMealID(null);
    }
  }

  async function onDeleteMeal(mealEventID: number) {
    const confirmed = window.confirm("Remove this meal from your log?");
    if (!confirmed) {
      return;
    }
    setDeletingMealID(mealEventID);
    setEditMealError(null);
    try {
      await deleteMeal(mealEventID, DASHBOARD_USER_ID);
      if (editingMealID === mealEventID) {
        setEditingMealID(null);
        setEditingMealValue("");
      }
      await refreshDashboard();
    } catch (error) {
      setEditMealError(error instanceof Error ? error.message : "Could not delete meal");
    } finally {
      setDeletingMealID(null);
    }
  }

  return (
    <main className="mx-auto w-full max-w-6xl space-y-5 px-4 pb-12 pt-4 sm:space-y-6 sm:px-6 lg:space-y-7 lg:px-8">
      <section className="rounded-[1.75rem] border border-stone-200/80 bg-gradient-to-b from-stone-50 via-amber-50/50 to-rose-50/20 p-5 shadow-[0_10px_28px_-16px_rgba(120,113,108,0.35)] sm:p-7">
        <p className="text-[11px] font-medium uppercase tracking-[0.16em] text-stone-500">Today’s Goals & Totals</p>
        <h2 className="mt-2 text-2xl font-semibold tracking-tight text-stone-900 sm:text-3xl lg:text-[2.05rem]">Nutrition Dashboard</h2>
        <p className="mt-1.5 max-w-2xl text-sm leading-relaxed text-stone-600 sm:text-base">Quick progress snapshot against your daily targets.</p>
        {dashboardError && <p className="mt-2 text-xs text-rose-700">{dashboardError}</p>}
        {isDashboardLoading && <p className="mt-2 text-xs text-stone-500">Loading latest totals…</p>}
        <div className="mt-5 grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 lg:grid-cols-4">
          {goals.map((goal) => (
            <GoalCard key={goal.label} {...goal} />
          ))}
        </div>
      </section>

      <section className="rounded-[1.75rem] border border-amber-200/70 bg-gradient-to-b from-white via-amber-50/65 to-orange-50/30 p-5 shadow-[0_10px_24px_-16px_rgba(217,119,6,0.35)] sm:p-7">
        <div className="mb-4 space-y-1.5">
          <p className="text-[11px] font-medium uppercase tracking-[0.16em] text-amber-800/80">Main Chat</p>
          <h3 className="text-2xl font-semibold tracking-tight text-stone-900 sm:text-[1.7rem]">Tell me what you need</h3>
          <p className="max-w-2xl text-sm leading-relaxed text-stone-600 sm:text-base">
            Log meals, track your weight, ask for recommendations, or just chat about your nutrition day.
          </p>
        </div>

        <form onSubmit={onSubmitChat} className="space-y-3">
          <label htmlFor="chat-input" className="sr-only">
            Main nutrition chat input
          </label>
          <div className="rounded-2xl border border-stone-300/90 bg-white/95 p-2.5 shadow-[0_8px_24px_-18px_rgba(41,37,36,0.55)] focus-within:border-amber-400 sm:p-3">
            <textarea
              id="chat-input"
              value={chatInput}
              onChange={(event) => setChatInput(event.target.value)}
              rows={3}
              placeholder="Log a meal, your weight, or ask what to eat next"
              className="w-full resize-none rounded-xl border-none bg-transparent px-2 py-1 text-sm leading-relaxed text-stone-800 outline-none placeholder:text-stone-400 sm:text-base"
            />
            <div className="flex flex-col items-start justify-between gap-2 px-2 pb-1 pt-2 sm:flex-row sm:items-center sm:gap-3">
              <p className="text-xs text-stone-500 sm:text-sm">Try: “Lunch was chicken bowl” or “I weigh 176.2 lb”</p>
              <button
                type="submit"
                disabled={isSubmitting}
                className="shrink-0 rounded-xl bg-stone-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-stone-700 disabled:opacity-70 sm:px-5"
              >
                {isSubmitting ? "Sending..." : "Send"}
              </button>
            </div>
          </div>
        </form>

        {chatError && <p className="mt-2 text-xs text-rose-700">{chatError}</p>}

        {chatTurns.length > 0 && (
          <div className="mt-4 space-y-3.5">
            {chatTurns.map((turn) => (
              <article key={turn.id} className="space-y-2.5">
                <div className="flex justify-end">
                  <p className="max-w-[92%] rounded-2xl bg-stone-900 px-3.5 py-2.5 text-sm leading-relaxed text-white shadow-[0_8px_18px_-14px_rgba(28,25,23,0.85)] sm:max-w-[75%]">
                    {turn.userMessage}
                  </p>
                </div>
                <div className="flex justify-start">
                  <div className="w-full max-w-[94%] rounded-2xl border border-stone-200 bg-white/95 px-3.5 py-3 text-sm text-stone-700 shadow-[0_10px_20px_-16px_rgba(41,37,36,0.55)] sm:max-w-[80%]">
                    <ChatResponseBlock response={turn.response} />
                  </div>
                </div>
              </article>
            ))}
          </div>
        )}
      </section>

      <SectionCard title={formatDashboardDate(dashboardDate)} subtitle="Meals logged for this date">
        {mealsError && <p className="mb-3 text-xs text-rose-700">{mealsError}</p>}
        {editMealError && <p className="mb-3 text-xs text-rose-700">{editMealError}</p>}
        {recentMeals.length === 0 ? (
          <div className="rounded-xl border border-stone-200 bg-white/90 px-4 py-6 text-center text-sm text-stone-500">
            No meals logged for this date — try logging one in chat above.
          </div>
        ) : (
          <div className="space-y-3">
            {recentMeals.map((meal) => {
              const timeHint = getMealTimeHint(meal.time_source);
              return (
                <article
                  key={meal.meal_event_id}
                  className="rounded-2xl border border-stone-200/90 bg-white/95 px-3.5 py-3.5 shadow-[0_10px_20px_-16px_rgba(41,37,36,0.5)] sm:px-4"
                >
                  <div className="sm:flex sm:items-center sm:justify-between">
                    <div>
                      <p className="text-[11px] font-medium uppercase tracking-[0.1em] text-stone-500">{formatMealTime(meal.eaten_at)}</p>
                      {timeHint && <p className="mt-0.5 text-[11px] text-stone-500">{timeHint}</p>}
                      <p className="mt-1 font-medium text-stone-900">{meal.canonical_name}</p>
                      {editingMealID === meal.meal_event_id ? (
                        <div className="mt-2 flex flex-wrap items-center gap-2">
                          <input
                            type="datetime-local"
                            value={editingMealValue}
                            onChange={(event) => setEditingMealValue(event.target.value)}
                            className="rounded-md border border-stone-300 bg-white px-2 py-1 text-xs text-stone-700"
                          />
                          <button
                            type="button"
                            onClick={() => void onSaveMealTime(meal.meal_event_id)}
                            disabled={savingMealID === meal.meal_event_id || !editingMealValue}
                            className="rounded-md bg-stone-900 px-2 py-1 text-xs font-medium text-white"
                          >
                            {savingMealID === meal.meal_event_id ? "Saving..." : "Save"}
                          </button>
                          <button
                            type="button"
                            onClick={() => {
                              setEditingMealID(null);
                              setEditingMealValue("");
                            }}
                            className="text-xs text-stone-600"
                          >
                            Cancel
                          </button>
                        </div>
                      ) : (
                        <div className="mt-1 flex flex-wrap items-center gap-3">
                          <button
                            type="button"
                            onClick={() => {
                              setEditingMealID(meal.meal_event_id);
                              setEditingMealValue(toDateTimeLocalValue(meal.eaten_at));
                              setEditMealError(null);
                            }}
                            className="text-xs font-medium text-stone-600 underline decoration-stone-300 underline-offset-2"
                          >
                            Edit time
                          </button>
                          <button
                            type="button"
                            onClick={() => void onDeleteMeal(meal.meal_event_id)}
                            disabled={deletingMealID === meal.meal_event_id}
                            className="text-xs font-medium text-rose-700 underline decoration-rose-300 underline-offset-2 disabled:opacity-60"
                          >
                            {deletingMealID === meal.meal_event_id ? "Removing..." : "Remove"}
                          </button>
                        </div>
                      )}
                    </div>
                    <p className="mt-2 inline-flex rounded-full border border-amber-200/80 bg-amber-50 px-2.5 py-1 text-xs font-medium text-amber-900 sm:mt-0">
                      {toNumber(meal.calories_kcal)} kcal
                    </p>
                  </div>
                  <div className="mt-3 grid grid-cols-2 gap-2.5 text-xs sm:grid-cols-4">
                    <MacroPill label="Protein" value={`${toNumber(meal.protein_g)}g`} />
                    <MacroPill label="Carbs" value={`${toNumber(meal.carbohydrate_g)}g`} />
                    <MacroPill label="Fat" value={`${toNumber(meal.fat_g)}g`} />
                    <MacroPill label="Calories" value={`${toNumber(meal.calories_kcal)}`} />
                  </div>
                </article>
              );
            })}
          </div>
        )}
      </SectionCard>

      <SectionCard title="Trends" subtitle="Select a metric and time range to view trend direction">
        <div className="space-y-3 rounded-2xl border border-stone-200/90 bg-white/95 p-3.5 shadow-[0_10px_20px_-16px_rgba(41,37,36,0.45)] sm:p-4">
          <div className="flex flex-wrap gap-2">
            {(Object.keys(metricLabels) as TrendMetric[]).map((metric) => (
              <button
                key={metric}
                type="button"
                onClick={() => setActiveMetric(metric)}
                className={`rounded-full px-3 py-1.5 text-xs font-medium transition sm:text-sm ${
                  activeMetric === metric
                    ? "bg-stone-900 text-white shadow-sm"
                    : "border border-stone-300 bg-stone-50 text-stone-700 hover:bg-stone-100"
                }`}
              >
                {metricLabels[metric]}
              </button>
            ))}
          </div>

          <div className="flex flex-wrap gap-2">
            {(Object.keys(rangeLabels) as TrendRange[]).map((range) => (
              <button
                key={range}
                type="button"
                onClick={() => setActiveRange(range)}
                className={`rounded-full px-3 py-1.5 text-xs transition sm:text-sm ${
                  activeRange === range
                    ? "border border-amber-300 bg-amber-100 text-amber-900 shadow-sm"
                    : "border border-stone-300 bg-white text-stone-600 hover:bg-stone-50"
                }`}
              >
                {rangeLabels[range]}
              </button>
            ))}
          </div>

          <LineTrendChart
            points={trendPointsByRange[activeRange] ?? []}
            metric={activeMetric}
            loading={isTrendsLoading}
            error={trendsError}
          />
        </div>
      </SectionCard>
    </main>
  );
}

function LineTrendChart({
  points,
  metric,
  loading,
  error
}: {
  points: TrendsPoint[];
  metric: TrendMetric;
  loading: boolean;
  error: string | null;
}) {
  const chartPoints = points
    .map((item) => {
      let value: number | null;
      if (metric === "calories") return item.calories_kcal;
      if (metric === "protein") return item.protein_g;
      if (metric === "carbs") return item.carbohydrate_g;
      if (metric === "fat") return item.fat_g;
      value = item.weight;
      return value;
    })
    .map((value, idx) => ({ value, date: points[idx]?.date }))
    .filter((item): item is { value: number; date: string } => item.value !== null && Boolean(item.date));

  const values = chartPoints.map((item) => item.value);

  if (error) {
    return <p className="text-sm text-rose-700">{error}</p>;
  }

  if (loading) {
    return <p className="text-sm text-stone-500">Loading trends…</p>;
  }

  if (values.length === 0 || points.length === 0) {
    return (
      <div className="rounded-2xl border border-stone-200 bg-stone-50/80 px-4 py-6 text-center text-sm text-stone-500">
        No trend data yet for this range.
      </div>
    );
  }

  const min = Math.min(...values);
  const max = Math.max(...values);
  const spread = max - min || 1;
  const chartWidth = 100;
  const chartHeight = 50;

  const linePoints = values
    .map((value, index) => {
      const x = (index / Math.max(values.length - 1, 1)) * chartWidth;
      const y = chartHeight - ((value - min) / spread) * chartHeight;
      return `${x},${y}`;
    })
    .join(" ");

  const latestValue = values.at(-1);
  const prevValue = values.at(-2) ?? latestValue;
  const delta = latestValue !== undefined && prevValue !== undefined ? latestValue - prevValue : 0;
  const deltaPrefix = delta >= 0 ? "+" : "";

  return (
    <div className="rounded-2xl border border-stone-200/90 bg-gradient-to-b from-stone-50 via-amber-50/30 to-white p-3.5 sm:p-4">
      <div className="mb-2.5 flex items-end justify-between gap-3">
        <div>
          <p className="text-[11px] uppercase tracking-[0.12em] text-stone-500">{metricLabels[metric]}</p>
          <p className="text-xl font-semibold tracking-tight text-stone-900">
            {latestValue}
            <span className="ml-1 text-sm font-medium text-stone-500">{metric === "weight" ? "lb" : metric === "calories" ? "kcal" : "g"}</span>
          </p>
        </div>
        <p className={`text-xs font-medium ${delta >= 0 ? "text-amber-800" : "text-emerald-700"}`}>
          {deltaPrefix}
          {delta.toFixed(metric === "weight" ? 1 : 0)} vs prior
        </p>
      </div>

      <div className="relative h-44 w-full rounded-xl border border-stone-200 bg-white p-2.5">
        <svg viewBox="0 0 100 50" preserveAspectRatio="none" className="h-full w-full">
          <defs>
            <linearGradient id="trend-fill" x1="0" x2="0" y1="0" y2="1">
              <stop offset="0%" stopColor="#f59e0b" stopOpacity="0.26" />
              <stop offset="100%" stopColor="#f59e0b" stopOpacity="0.03" />
            </linearGradient>
          </defs>
          <polyline fill="none" stroke="#f59e0b" strokeWidth="1.5" points={linePoints} />
          <polygon fill="url(#trend-fill)" points={`0,50 ${linePoints} 100,50`} />
        </svg>
        <div className="mt-2 grid grid-cols-4 gap-1 text-[10px] text-stone-500 sm:grid-cols-6 md:grid-cols-8">
          {chartPoints.filter((_, idx) => idx % Math.ceil(chartPoints.length / 8) === 0 || idx === chartPoints.length - 1).map((point) => (
            <span key={point.date} className="truncate text-center">
              {formatTrendLabel(point.date)}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}

function MacroPill({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-stone-200 bg-stone-50/80 px-2.5 py-2.5">
      <p className="text-[10px] uppercase tracking-[0.1em] text-stone-500">{label}</p>
      <p className="mt-1 text-sm font-semibold tracking-tight text-stone-800">{value}</p>
    </div>
  );
}

function toNumber(value: number | null | undefined) {
  return Math.round(value ?? 0);
}

function formatMealTime(timestamp: string) {
  const parsed = new Date(timestamp);
  if (Number.isNaN(parsed.getTime())) {
    return "--:--";
  }
  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit"
  }).format(parsed);
}

function getMealTimeHint(timeSource?: string | null) {
  if (!timeSource) {
    return null;
  }
  if (timeSource === "default_now") {
    return "Logged for now. Edit time if needed.";
  }
  return null;
}

function toDateTimeLocalValue(timestamp: string) {
  const parsed = new Date(timestamp);
  if (Number.isNaN(parsed.getTime())) {
    return "";
  }
  const pad = (value: number) => String(value).padStart(2, "0");
  return `${parsed.getFullYear()}-${pad(parsed.getMonth() + 1)}-${pad(parsed.getDate())}T${pad(parsed.getHours())}:${pad(parsed.getMinutes())}`;
}

function ChatResponseBlock({ response }: { response: ChatResponse }) {
  if (response.meal_result) {
    return (
      <div className="space-y-2.5">
        <p className="text-[11px] font-medium uppercase tracking-[0.1em] text-amber-700">Meal Logged</p>
        <p className="text-[15px] font-semibold tracking-tight text-stone-900">{response.meal_result.canonical_name}</p>
        <div className="grid grid-cols-2 gap-2 text-xs sm:grid-cols-4">
          <MacroPill label="Calories" value={`${toNumber(response.meal_result.calories_kcal ?? null)}`} />
          <MacroPill label="Protein" value={`${toNumber(response.meal_result.protein_g ?? null)}g`} />
          <MacroPill label="Carbs" value={`${toNumber(response.meal_result.carbohydrate_g ?? null)}g`} />
          <MacroPill label="Fat" value={`${toNumber(response.meal_result.fat_g ?? null)}g`} />
        </div>
      </div>
    );
  }

  if (response.weight_result) {
    return (
      <div className="space-y-1.5">
        <p className="text-[11px] font-medium uppercase tracking-[0.1em] text-amber-700">Weight Logged</p>
        <p className="text-base font-semibold tracking-tight text-stone-900">
          {response.weight_result.weight} {response.weight_result.unit}
        </p>
      </div>
    );
  }

  if (response.recommendation_result) {
    const recommendations = toRecommendationList(response.recommendation_result.text);
    return (
      <div className="space-y-2.5">
        <p className="text-[11px] font-medium uppercase tracking-[0.1em] text-amber-700">Recommendations</p>
        <ul className="space-y-1 text-sm text-stone-700">
          {recommendations.map((item, idx) => (
            <li key={`${item}-${idx}`} className="flex gap-2">
              <span className="mt-1 h-1.5 w-1.5 rounded-full bg-amber-400" />
              <span>{item}</span>
            </li>
          ))}
        </ul>
      </div>
    );
  }

  return <p className="text-sm text-stone-700">{response.message_to_user || "Thanks for your message."}</p>;
}

function currentLocalISODate() {
  return toLocalISODate(new Date());
}

function toLocalISODate(date: Date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function toRecommendationList(text: string) {
  const raw = text
    .split(/\n|;/)
    .map((part) => part.replace(/^[-•\s]+/, "").trim())
    .filter(Boolean);
  if (raw.length > 1) {
    return raw;
  }

  const commaSplit = text
    .split(",")
    .map((part) => part.trim())
    .filter((part) => part.length > 0);
  if (commaSplit.length > 1) {
    return commaSplit;
  }

  return [text.trim()].filter(Boolean);
}

function isMealOnDate(timestamp: string, isoDate: string) {
  const parsed = new Date(timestamp);
  if (Number.isNaN(parsed.getTime())) {
    return false;
  }
  return toLocalISODate(parsed) === isoDate;
}

function formatDashboardDate(isoDate: string) {
  const parsed = new Date(`${isoDate}T12:00:00`);
  if (Number.isNaN(parsed.getTime())) {
    return "Meals";
  }
  return new Intl.DateTimeFormat("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric",
    year: "numeric"
  }).format(parsed);
}

function formatTrendLabel(date: string) {
  const parsed = new Date(date);
  if (Number.isNaN(parsed.getTime())) {
    return date;
  }
  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric"
  }).format(parsed);
}

function SectionCard({
  title,
  subtitle,
  children
}: {
  title: string;
  subtitle: string;
  children: React.ReactNode;
}) {
  return (
    <section className="rounded-[1.6rem] border border-stone-200/80 bg-gradient-to-b from-stone-50/90 to-white p-5 shadow-[0_10px_24px_-18px_rgba(41,37,36,0.5)] sm:p-6">
      <header className="mb-4">
        <h3 className="text-xl font-semibold tracking-tight text-stone-900 sm:text-[1.45rem]">{title}</h3>
        <p className="mt-1 text-sm leading-relaxed text-stone-600 sm:text-base">{subtitle}</p>
      </header>
      {children}
    </section>
  );
}

function GoalCard({
  label,
  consumed,
  target,
  unit
}: {
  label: string;
  consumed: number;
  target: number;
  unit: string;
}) {
  const pct = Math.min(Math.round((consumed / target) * 100), 100);
  const remaining = Math.max(target - consumed, 0);

  return (
    <article className="rounded-2xl border border-stone-200/90 bg-white/95 p-3.5 shadow-[0_8px_18px_-14px_rgba(41,37,36,0.45)]">
      <p className="text-[11px] font-medium uppercase tracking-[0.12em] text-stone-500">{label}</p>
      <p className="mt-1.5 text-xl font-semibold tracking-tight text-stone-900">
        {consumed}
        <span className="ml-1 text-sm font-medium text-stone-500">/ {target}</span>
      </p>
      <p className="text-xs text-stone-600">{remaining} {unit} remaining</p>
      <div className="mt-3 h-2.5 rounded-full bg-stone-200">
        <div className="h-2.5 rounded-full bg-amber-300" style={{ width: `${pct}%` }} />
      </div>
      <p className="mt-1.5 text-[11px] text-stone-500">{pct}% complete</p>
    </article>
  );
}
