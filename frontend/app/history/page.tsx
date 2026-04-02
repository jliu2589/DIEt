"use client";

import { useMemo, useState } from "react";

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

const MOCK_MEAL_CATALOG: JournalMeal[] = [
  { id: "a", time: "8:10 AM", name: "Greek yogurt, berries, granola", calories: 430, protein: 28, carbs: 47, fat: 14 },
  { id: "b", time: "12:40 PM", name: "Chicken rice bowl", calories: 620, protein: 42, carbs: 58, fat: 22 },
  { id: "c", time: "4:20 PM", name: "Protein shake", calories: 280, protein: 32, carbs: 18, fat: 7 },
  { id: "d", time: "7:35 PM", name: "Salmon, potatoes, salad", calories: 710, protein: 46, carbs: 52, fat: 30 },
  { id: "e", time: "9:10 PM", name: "Dark chocolate + almonds", calories: 220, protein: 6, carbs: 14, fat: 16 }
];

function makeSevenDayJournal(rangeEndDate: Date): JournalDay[] {
  const days: JournalDay[] = [];

  for (let offset = 6; offset >= 0; offset -= 1) {
    const date = new Date(rangeEndDate);
    date.setHours(0, 0, 0, 0);
    date.setDate(date.getDate() - offset);

    const count = ((date.getDate() + date.getMonth()) % 3) + 2; // 2-4 meals/day
    const meals = Array.from({ length: count }, (_, idx) => {
      const seed = (date.getDate() + idx + date.getMonth()) % MOCK_MEAL_CATALOG.length;
      const base = MOCK_MEAL_CATALOG[seed];
      return {
        ...base,
        id: `${date.toISOString()}-${idx}-${base.id}`
      };
    });

    const totals = meals.reduce(
      (acc, meal) => ({
        calories: acc.calories + meal.calories,
        protein: acc.protein + meal.protein,
        carbs: acc.carbs + meal.carbs,
        fat: acc.fat + meal.fat
      }),
      { calories: 0, protein: 0, carbs: 0, fat: 0 }
    );

    days.push({
      isoDate: date.toISOString().slice(0, 10),
      meals,
      totals
    });
  }

  return days;
}

function formatSevenDayRange(days: JournalDay[]) {
  if (days.length === 0) {
    return "7-day journal";
  }
  const first = new Date(`${days[0].isoDate}T00:00:00`);
  const last = new Date(`${days[days.length - 1].isoDate}T00:00:00`);

  const start = new Intl.DateTimeFormat("en-US", { month: "long", day: "numeric" }).format(first);
  const end = new Intl.DateTimeFormat("en-US", { month: "short", day: "numeric" }).format(last);
  return `${start} – ${end}`;
}

function formatDayLabel(isoDate: string) {
  const date = new Date(`${isoDate}T00:00:00`);
  return new Intl.DateTimeFormat("en-US", {
    weekday: "long",
    month: "short",
    day: "numeric"
  }).format(date);
}

function MacroTile({ label, value, unit }: { label: string; value: number; unit: string }) {
  return (
    <div className="rounded-xl border border-stone-200/90 bg-stone-50/85 px-3 py-2">
      <p className="text-[10px] uppercase tracking-[0.12em] text-stone-500">{label}</p>
      <p className="mt-1 text-sm font-semibold tracking-tight text-stone-900">
        {value}
        <span className="ml-1 text-xs font-medium text-stone-500">{unit}</span>
      </p>
    </div>
  );
}

function DayJournalSection({ day }: { day: JournalDay }) {
  return (
    <section className="rounded-[1.45rem] border border-stone-200/85 bg-gradient-to-b from-white to-stone-50/85 p-4 shadow-[0_10px_24px_-18px_rgba(41,37,36,0.4)] sm:p-5">
      <header className="mb-3.5 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
        <h2 className="text-lg font-semibold tracking-tight text-stone-900 sm:text-xl">{formatDayLabel(day.isoDate)}</h2>
        <p className="text-xs text-stone-500">{day.meals.length} meals logged</p>
      </header>

      <div className="mb-4 grid grid-cols-2 gap-2.5 sm:grid-cols-4">
        <MacroTile label="Calories" value={day.totals.calories} unit="kcal" />
        <MacroTile label="Protein" value={day.totals.protein} unit="g" />
        <MacroTile label="Carbs" value={day.totals.carbs} unit="g" />
        <MacroTile label="Fat" value={day.totals.fat} unit="g" />
      </div>

      <div className="space-y-2.5">
        <div className="hidden rounded-xl border border-stone-200/90 bg-stone-50/75 px-3 py-2 text-[11px] font-medium uppercase tracking-[0.08em] text-stone-500 md:grid md:grid-cols-[120px_1fr_90px_80px_80px_70px]">
          <p>Time</p>
          <p>Meal</p>
          <p>Calories</p>
          <p>Protein</p>
          <p>Carbs</p>
          <p>Fat</p>
        </div>

        {day.meals.map((meal) => (
          <article key={meal.id} className="rounded-xl border border-stone-200/90 bg-white/95 px-3.5 py-3 shadow-[0_8px_16px_-14px_rgba(41,37,36,0.45)]">
            <div className="md:hidden">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="text-[11px] font-medium uppercase tracking-[0.1em] text-stone-500">{meal.time}</p>
                  <p className="mt-1 text-sm font-medium text-stone-900">{meal.name}</p>
                </div>
                <p className="inline-flex rounded-full border border-amber-200/80 bg-amber-50 px-2.5 py-1 text-xs font-medium text-amber-900">
                  {meal.calories} kcal
                </p>
              </div>
              <div className="mt-3 grid grid-cols-3 gap-2 text-xs text-stone-700">
                <p>Protein: {meal.protein}g</p>
                <p>Carbs: {meal.carbs}g</p>
                <p>Fat: {meal.fat}g</p>
              </div>
            </div>

            <div className="hidden items-center gap-3 text-sm text-stone-800 md:grid md:grid-cols-[120px_1fr_90px_80px_80px_70px]">
              <p className="text-xs font-medium uppercase tracking-[0.08em] text-stone-500">{meal.time}</p>
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
    now.setHours(0, 0, 0, 0);
    return now;
  });

  const journalDays = useMemo(() => makeSevenDayJournal(rangeEndDate), [rangeEndDate]);
  const rangeLabel = useMemo(() => formatSevenDayRange(journalDays), [journalDays]);

  function shiftWeek(daysDelta: number) {
    setRangeEndDate((prev) => {
      const next = new Date(prev);
      next.setDate(prev.getDate() + daysDelta);
      return next;
    });
  }

  return (
    <main className="mx-auto w-full max-w-5xl space-y-5 px-4 pb-12 pt-4 sm:space-y-6 sm:px-6 lg:px-8">
      <section className="rounded-[1.75rem] border border-stone-200/85 bg-gradient-to-b from-stone-50 via-amber-50/45 to-rose-50/20 p-5 shadow-[0_10px_28px_-16px_rgba(120,113,108,0.35)] sm:p-7">
        <p className="text-[11px] font-medium uppercase tracking-[0.16em] text-stone-500">History</p>
        <div className="mt-3 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight text-stone-900 sm:text-3xl">7-day nutrition journal</h1>
            <p className="mt-2 text-xl font-medium tracking-tight text-stone-800 sm:text-2xl">{rangeLabel}</p>
            <p className="mt-1 text-sm leading-relaxed text-stone-600">Daily totals and exact meals, grouped by day.</p>
          </div>
          <div className="flex items-center gap-2 self-start sm:self-auto">
            <button
              type="button"
              onClick={() => shiftWeek(-7)}
              className="rounded-full border border-stone-300/90 bg-white/90 px-3 py-1.5 text-xs font-medium text-stone-700 transition hover:bg-stone-100"
            >
              ← Previous
            </button>
            <button
              type="button"
              onClick={() => shiftWeek(7)}
              className="rounded-full border border-stone-300/90 bg-white/90 px-3 py-1.5 text-xs font-medium text-stone-700 transition hover:bg-stone-100"
            >
              Next →
            </button>
          </div>
        </div>
      </section>

      <div className="space-y-4 sm:space-y-5">
        {journalDays.map((day) => (
          <DayJournalSection key={day.isoDate} day={day} />
        ))}
      </div>
    </main>
  );
}
