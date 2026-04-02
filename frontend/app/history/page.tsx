"use client";

import { useEffect, useMemo, useState } from "react";
import { getDailySummary, getRecentMeals, type DailySummaryResponse, type RecentMeal } from "@/lib/api";

type JournalMeal = {
  id: string;
  time: string;
  name: string;
  calories: number;
  protein: number;
  carbs: number;
  fat: number;
};

type JournalDay = {
  isoDate: string;
  meals: JournalMeal[];
  totals: {
    calories: number;
    protein: number;
    carbs: number;
    fat: number;
  };
};

const USER_ID = "demo-user-001";

function toNumber(value: number | null | undefined) {
  return Math.round(value ?? 0);
}

function formatMealTime(value: string) {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return "—";
  }
  return new Intl.DateTimeFormat("en-US", {
    hour: "numeric",
    minute: "2-digit"
  }).format(parsed);
}

function getISODateUTC(date: Date) {
  const copy = new Date(date);
  copy.setUTCHours(0, 0, 0, 0);
  return copy.toISOString().slice(0, 10);
}

function buildSevenDayWindow(rangeEndDate: Date) {
  const dates: string[] = [];
  for (let offset = 6; offset >= 0; offset -= 1) {
    const d = new Date(rangeEndDate);
    d.setUTCDate(d.getUTCDate() - offset);
    dates.push(getISODateUTC(d));
  }
  return dates;
}

function formatSevenDayRange(days: JournalDay[]) {
  if (days.length === 0) {
    return "7-day journal";
  }
  const first = new Date(`${days[0].isoDate}T00:00:00Z`);
  const last = new Date(`${days[days.length - 1].isoDate}T00:00:00Z`);

  const start = new Intl.DateTimeFormat("en-US", { month: "long", day: "numeric" }).format(first);
  const end = new Intl.DateTimeFormat("en-US", { month: "short", day: "numeric" }).format(last);
  return `${start} – ${end}`;
}

function formatDayLabel(isoDate: string) {
  const date = new Date(`${isoDate}T00:00:00Z`);
  return new Intl.DateTimeFormat("en-US", {
    weekday: "long",
    month: "short",
    day: "numeric"
  }).format(date);
}

function MacroTile({ label, value, unit }: { label: string; value: number; unit: string }) {
  return (
    <div className="rounded-2xl border border-stone-200/80 bg-white/80 px-3.5 py-2.5 shadow-[0_8px_20px_-18px_rgba(41,37,36,0.45)] backdrop-blur-[1px] sm:px-4">
      <p className="text-[10px] uppercase tracking-[0.14em] text-stone-500">{label}</p>
      <p className="mt-1 text-base font-semibold tracking-tight text-stone-900">
        {value}
        <span className="ml-1 text-xs font-medium text-stone-500">{unit}</span>
      </p>
    </div>
  );
}

function DayJournalSection({ day }: { day: JournalDay }) {
  return (
    <section className="rounded-[1.6rem] border border-stone-200/80 bg-gradient-to-b from-white via-stone-50/55 to-amber-50/25 p-4 shadow-[0_18px_32px_-28px_rgba(41,37,36,0.45)] sm:p-6">
      <header className="mb-4 flex flex-col gap-1.5 border-b border-stone-200/70 pb-3.5 sm:mb-5 sm:flex-row sm:items-end sm:justify-between sm:pb-4">
        <h2 className="text-xl font-semibold tracking-tight text-stone-900 sm:text-[1.38rem]">{formatDayLabel(day.isoDate)}</h2>
        <p className="text-xs font-medium tracking-[0.04em] text-stone-500">{day.meals.length} meals logged</p>
      </header>

      <div className="mb-4 grid grid-cols-2 gap-2.5 sm:mb-5 sm:grid-cols-4 sm:gap-3">
        <MacroTile label="Calories" value={day.totals.calories} unit="kcal" />
        <MacroTile label="Protein" value={day.totals.protein} unit="g" />
        <MacroTile label="Carbs" value={day.totals.carbs} unit="g" />
        <MacroTile label="Fat" value={day.totals.fat} unit="g" />
      </div>

      <div className="space-y-2.5 sm:space-y-3">
        {day.meals.length === 0 && (
          <div className="rounded-2xl border border-stone-200/80 bg-white/75 px-4 py-5 text-sm text-stone-600">No meals logged.</div>
        )}
        {day.meals.length > 0 && (
          <div className="hidden rounded-2xl border border-stone-200/80 bg-white/70 px-3.5 py-2.5 text-[11px] font-medium uppercase tracking-[0.1em] text-stone-500 md:grid md:grid-cols-[110px_1fr_90px_80px_80px_70px]">
            <p>Time</p>
            <p>Meal</p>
            <p>Calories</p>
            <p>Protein</p>
            <p>Carbs</p>
            <p>Fat</p>
          </div>
        )}

        {day.meals.map((meal) => (
          <article key={meal.id} className="rounded-2xl border border-stone-200/80 bg-white/90 px-3.5 py-3.5 shadow-[0_12px_22px_-18px_rgba(41,37,36,0.45)] sm:px-4">
            <div className="md:hidden">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="text-[11px] font-medium uppercase tracking-[0.11em] text-stone-500">{meal.time}</p>
                  <p className="mt-1.5 text-[15px] font-medium leading-snug text-stone-900">{meal.name}</p>
                </div>
                <p className="inline-flex rounded-full border border-amber-200/80 bg-amber-50/80 px-2.5 py-1 text-xs font-medium text-amber-900">{meal.calories} kcal</p>
              </div>
              <div className="mt-3.5 grid grid-cols-3 gap-2 rounded-xl border border-stone-200/70 bg-stone-50/70 px-2.5 py-2 text-xs text-stone-700">
                <p className="font-medium">P {meal.protein}g</p>
                <p className="font-medium">C {meal.carbs}g</p>
                <p className="font-medium">F {meal.fat}g</p>
              </div>
            </div>

            <div className="hidden items-center gap-3 text-sm text-stone-800 md:grid md:grid-cols-[110px_1fr_90px_80px_80px_70px]">
              <p className="text-xs font-medium uppercase tracking-[0.1em] text-stone-500">{meal.time}</p>
              <p className="font-medium text-stone-900">{meal.name}</p>
              <p>{meal.calories} kcal</p>
              <p>{meal.protein}g</p>
              <p>{meal.carbs}g</p>
              <p>{meal.fat}g</p>
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}

export default function HistoryPage() {
  const [rangeEndDate, setRangeEndDate] = useState(() => {
    const now = new Date();
    now.setUTCHours(0, 0, 0, 0);
    return now;
  });
  const [journalDays, setJournalDays] = useState<JournalDay[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const rangeLabel = useMemo(() => formatSevenDayRange(journalDays), [journalDays]);
  const hasAnyMeals = useMemo(() => journalDays.some((day) => day.meals.length > 0), [journalDays]);

  function shiftWeek(daysDelta: number) {
    setRangeEndDate((prev) => {
      const next = new Date(prev);
      next.setUTCDate(prev.getUTCDate() + daysDelta);
      return next;
    });
  }

  useEffect(() => {
    async function loadJournal() {
      setLoading(true);
      setError(null);

      const dates = buildSevenDayWindow(rangeEndDate);

      try {
        const [recentMealsResult, summaries] = await Promise.all([
          getRecentMeals(USER_ID, 200),
          Promise.all(
            dates.map(async (date) => {
              try {
                return await getDailySummary(USER_ID, date);
              } catch {
                return null;
              }
            })
          )
        ]);

        const groupedMeals = new Map<string, JournalMeal[]>();
        for (const date of dates) {
          groupedMeals.set(date, []);
        }

        for (const meal of recentMealsResult.items ?? []) {
          const mealDate = meal.eaten_at.slice(0, 10);
          if (!groupedMeals.has(mealDate)) {
            continue;
          }
          groupedMeals.get(mealDate)?.push({
            id: String(meal.meal_event_id),
            time: formatMealTime(meal.eaten_at),
            name: meal.canonical_name,
            calories: toNumber(meal.calories_kcal),
            protein: toNumber(meal.protein_g),
            carbs: toNumber(meal.carbohydrate_g),
            fat: toNumber(meal.fat_g)
          });
        }

        const summaryByDate = new Map<string, DailySummaryResponse>();
        summaries.forEach((summary) => {
          if (summary?.date) {
            summaryByDate.set(summary.date, summary);
          }
        });

        const nextDays = dates.map((date) => {
          const meals = groupedMeals.get(date) ?? [];
          const summary = summaryByDate.get(date);

          const fallbackTotals = meals.reduce(
            (acc, meal) => ({
              calories: acc.calories + meal.calories,
              protein: acc.protein + meal.protein,
              carbs: acc.carbs + meal.carbs,
              fat: acc.fat + meal.fat
            }),
            { calories: 0, protein: 0, carbs: 0, fat: 0 }
          );

          return {
            isoDate: date,
            meals,
            totals: {
              calories: toNumber(summary?.totals?.calories_kcal ?? fallbackTotals.calories),
              protein: toNumber(summary?.totals?.protein_g ?? fallbackTotals.protein),
              carbs: toNumber(summary?.totals?.carbohydrate_g ?? fallbackTotals.carbs),
              fat: toNumber(summary?.totals?.fat_g ?? fallbackTotals.fat)
            }
          };
        });

        setJournalDays(nextDays);

        if (!summaries.some((s) => s !== null)) {
          setError("Some daily totals are unavailable right now. Showing meal-based estimates where needed.");
        }
      } catch {
        setError("We couldn't load your history right now. Please try again shortly.");
        setJournalDays(
          dates.map((date) => ({
            isoDate: date,
            meals: [],
            totals: { calories: 0, protein: 0, carbs: 0, fat: 0 }
          }))
        );
      } finally {
        setLoading(false);
      }
    }

    void loadJournal();
  }, [rangeEndDate]);

  return (
    <main className="mx-auto w-full max-w-5xl space-y-5 px-4 pb-14 pt-4 sm:space-y-6 sm:px-6 sm:pt-6 lg:px-8">
      <section className="rounded-[1.9rem] border border-stone-200/80 bg-gradient-to-b from-stone-50 via-amber-50/40 to-rose-50/15 p-5 shadow-[0_16px_34px_-22px_rgba(120,113,108,0.38)] sm:p-7">
        <p className="text-[11px] font-medium uppercase tracking-[0.18em] text-stone-500">History</p>
        <div className="mt-3.5 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight text-stone-900 sm:text-[2rem]">7-day nutrition journal</h1>
            <p className="mt-2 text-lg font-medium tracking-tight text-stone-800 sm:text-2xl">{rangeLabel || "7-day journal"}</p>
            <p className="mt-1.5 max-w-xl text-sm leading-relaxed text-stone-600">Daily totals and exact meals, grouped by day.</p>
          </div>
          <div className="flex w-full items-center gap-2.5 self-start sm:w-auto sm:self-auto">
            <button
              type="button"
              onClick={() => shiftWeek(-7)}
              className="flex-1 rounded-full border border-stone-300/90 bg-white/90 px-3.5 py-2 text-xs font-medium text-stone-700 transition hover:bg-stone-100 sm:flex-none"
            >
              ← Previous
            </button>
            <button
              type="button"
              onClick={() => shiftWeek(7)}
              className="flex-1 rounded-full border border-stone-300/90 bg-white/90 px-3.5 py-2 text-xs font-medium text-stone-700 transition hover:bg-stone-100 sm:flex-none"
            >
              Next →
            </button>
          </div>
        </div>
      </section>

      {error && <p className="rounded-2xl border border-amber-200/80 bg-amber-50/80 px-4 py-3 text-sm text-amber-900">{error}</p>}

      {loading ? (
        <section className="rounded-[1.6rem] border border-stone-200/80 bg-gradient-to-b from-white to-stone-50/80 p-6 text-center text-sm text-stone-600 shadow-[0_12px_26px_-20px_rgba(41,37,36,0.4)] sm:p-8">
          Loading your 7-day journal…
        </section>
      ) : !hasAnyMeals ? (
        <section className="rounded-[1.6rem] border border-stone-200/80 bg-gradient-to-b from-white to-stone-50/80 p-6 text-center shadow-[0_12px_26px_-20px_rgba(41,37,36,0.4)] sm:p-8">
          <p className="text-sm font-medium text-stone-700">No meals logged in this 7-day window yet.</p>
          <p className="mt-1 text-xs text-stone-500">Try another week, or start logging meals to build your journal.</p>
        </section>
      ) : (
        <div className="space-y-4 sm:space-y-5">
          {journalDays.map((day) => (
            <DayJournalSection key={day.isoDate} day={day} />
          ))}
        </div>
      )}
    </main>
  );
}
