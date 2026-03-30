"use client";

import { FormEvent, useState } from "react";

const goals = [
  { label: "Calories", consumed: 1380, target: 2500, unit: "kcal" },
  { label: "Protein", consumed: 92, target: 160, unit: "g" },
  { label: "Carbs", consumed: 145, target: 220, unit: "g" },
  { label: "Fat", consumed: 52, target: 75, unit: "g" }
];

const todaysMeals = [
  { id: 1, name: "Greek Yogurt + Berries", time: "8:10 AM", calories: 320, protein: 28, carbs: 32, fat: 9 },
  { id: 2, name: "Salmon Rice Bowl", time: "12:42 PM", calories: 610, protein: 41, carbs: 58, fat: 23 },
  { id: 3, name: "Almonds + Banana", time: "3:20 PM", calories: 210, protein: 9, carbs: 24, fat: 11 },
  { id: 4, name: "Chicken & Avocado Wrap", time: "7:05 PM", calories: 470, protein: 36, carbs: 34, fat: 21 }
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

const trendData: Record<TrendRange, Array<{ label: string; calories: number; protein: number; carbs: number; fat: number; weight: number }>> = {
  weekly: [
    { label: "Mon", calories: 1820, protein: 122, carbs: 188, fat: 61, weight: 176.6 },
    { label: "Tue", calories: 1960, protein: 130, carbs: 201, fat: 64, weight: 176.3 },
    { label: "Wed", calories: 1875, protein: 126, carbs: 174, fat: 60, weight: 176.1 },
    { label: "Thu", calories: 2050, protein: 136, carbs: 218, fat: 66, weight: 175.9 },
    { label: "Fri", calories: 1935, protein: 128, carbs: 196, fat: 63, weight: 175.8 },
    { label: "Sat", calories: 2120, protein: 140, carbs: 224, fat: 69, weight: 175.7 },
    { label: "Sun", calories: 1985, protein: 134, carbs: 209, fat: 65, weight: 175.6 }
  ],
  monthly: [
    { label: "W1", calories: 1900, protein: 126, carbs: 190, fat: 62, weight: 177.2 },
    { label: "W2", calories: 1940, protein: 129, carbs: 196, fat: 64, weight: 176.7 },
    { label: "W3", calories: 1880, protein: 124, carbs: 184, fat: 61, weight: 176.2 },
    { label: "W4", calories: 1960, protein: 132, carbs: 201, fat: 65, weight: 175.8 }
  ],
  "90d": [
    { label: "M1", calories: 2020, protein: 132, carbs: 214, fat: 68, weight: 179.4 },
    { label: "M2", calories: 1980, protein: 130, carbs: 207, fat: 66, weight: 178.6 },
    { label: "M3", calories: 1925, protein: 128, carbs: 201, fat: 64, weight: 177.7 },
    { label: "M4", calories: 1895, protein: 126, carbs: 194, fat: 63, weight: 176.9 },
    { label: "M5", calories: 1950, protein: 129, carbs: 199, fat: 64, weight: 176.1 },
    { label: "M6", calories: 1910, protein: 127, carbs: 192, fat: 62, weight: 175.8 }
  ],
  yearly: [
    { label: "Jan", calories: 2100, protein: 131, carbs: 226, fat: 71, weight: 181.1 },
    { label: "Feb", calories: 2040, protein: 129, carbs: 218, fat: 69, weight: 179.9 },
    { label: "Mar", calories: 1980, protein: 127, carbs: 209, fat: 66, weight: 178.7 },
    { label: "Apr", calories: 1950, protein: 128, carbs: 204, fat: 65, weight: 177.9 },
    { label: "May", calories: 1920, protein: 126, carbs: 198, fat: 64, weight: 177.2 },
    { label: "Jun", calories: 1890, protein: 125, carbs: 193, fat: 63, weight: 176.6 },
    { label: "Jul", calories: 1940, protein: 129, carbs: 199, fat: 64, weight: 176.1 },
    { label: "Aug", calories: 1900, protein: 126, carbs: 191, fat: 62, weight: 175.9 },
    { label: "Sep", calories: 1920, protein: 127, carbs: 194, fat: 63, weight: 175.8 },
    { label: "Oct", calories: 1880, protein: 124, carbs: 188, fat: 61, weight: 175.7 },
    { label: "Nov", calories: 1930, protein: 128, carbs: 197, fat: 64, weight: 175.6 },
    { label: "Dec", calories: 1890, protein: 125, carbs: 190, fat: 62, weight: 175.5 }
  ]
};

export default function HomePage() {
  const [chatInput, setChatInput] = useState("");
  const [drafts, setDrafts] = useState<string[]>([]);
  const [activeMetric, setActiveMetric] = useState<TrendMetric>("calories");
  const [activeRange, setActiveRange] = useState<TrendRange>("weekly");

  function onSubmitChat(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmed = chatInput.trim();
    if (!trimmed) {
      return;
    }
    setDrafts((prev) => [trimmed, ...prev].slice(0, 3));
    setChatInput("");
  }

  return (
    <main className="mx-auto w-full max-w-5xl space-y-6 px-4 pb-10 pt-3 sm:space-y-7 sm:px-6 lg:space-y-8 lg:px-8">
      <section className="rounded-3xl border border-stone-200/80 bg-gradient-to-b from-stone-50 to-amber-50/40 p-5 shadow-sm sm:p-7">
        <p className="text-xs font-medium uppercase tracking-[0.14em] text-stone-500">Today’s Goals & Totals</p>
        <h2 className="mt-2 text-2xl font-semibold tracking-tight text-stone-900 sm:text-3xl">Nutrition Dashboard</h2>
        <p className="mt-1 text-sm text-stone-600 sm:text-base">Quick progress snapshot against your daily targets.</p>
        <div className="mt-5 grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-4">
          {goals.map((goal) => (
            <GoalCard key={goal.label} {...goal} />
          ))}
        </div>
      </section>

      <section className="rounded-3xl border border-amber-200/70 bg-gradient-to-b from-white to-amber-50/70 p-5 shadow-sm sm:p-7">
        <div className="mb-4 space-y-1.5">
          <p className="text-xs font-medium uppercase tracking-[0.14em] text-amber-700/80">Main Chat</p>
          <h3 className="text-2xl font-semibold tracking-tight text-stone-900">Tell me what you need</h3>
          <p className="max-w-2xl text-sm text-stone-600 sm:text-base">
            Log meals, track your weight, ask for recommendations, or just chat about your nutrition day.
          </p>
        </div>

        <form onSubmit={onSubmitChat} className="space-y-3">
          <label htmlFor="chat-input" className="sr-only">
            Main nutrition chat input
          </label>
          <div className="rounded-2xl border border-stone-300 bg-white p-2 shadow-sm focus-within:border-amber-400 sm:p-3">
            <textarea
              id="chat-input"
              value={chatInput}
              onChange={(event) => setChatInput(event.target.value)}
              rows={3}
              placeholder="Log a meal, your weight, or ask what to eat next"
              className="w-full resize-none rounded-xl border-none bg-transparent px-2 py-1 text-sm leading-relaxed text-stone-800 outline-none placeholder:text-stone-400 sm:text-base"
            />
            <div className="flex items-center justify-between gap-3 px-2 pb-1 pt-2">
              <p className="text-xs text-stone-500 sm:text-sm">Try: “Lunch was chicken bowl” or “I weigh 176.2 lb”</p>
              <button
                type="submit"
                className="shrink-0 rounded-xl bg-stone-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-stone-700 sm:px-5"
              >
                Send
              </button>
            </div>
          </div>
        </form>

        {drafts.length > 0 && (
          <div className="mt-3 flex flex-wrap gap-2">
            {drafts.map((draft, idx) => (
              <p key={`${draft}-${idx}`} className="rounded-full border border-stone-200 bg-white px-3 py-1 text-xs text-stone-600">
                {draft}
              </p>
            ))}
          </div>
        )}
      </section>

      <SectionCard title="Today’s Meals" subtitle="Snapshot list with placeholder totals">
        <div className="space-y-2.5 sm:space-y-3">
          {todaysMeals.map((meal) => (
            <article
              key={meal.id}
              className="rounded-xl border border-stone-200 bg-white/95 px-3 py-3 shadow-sm sm:px-4"
            >
              <div className="sm:flex sm:items-center sm:justify-between">
                <div>
                  <p className="text-xs font-medium uppercase tracking-[0.08em] text-stone-500">{meal.time}</p>
                  <p className="mt-1 font-medium text-stone-900">{meal.name}</p>
                </div>
                <p className="mt-2 inline-flex rounded-full bg-amber-50 px-2.5 py-1 text-xs font-medium text-amber-800 sm:mt-0">
                  {meal.calories} kcal
                </p>
              </div>
              <div className="mt-3 grid grid-cols-2 gap-2 text-xs sm:grid-cols-4">
                <MacroPill label="Protein" value={`${meal.protein}g`} />
                <MacroPill label="Carbs" value={`${meal.carbs}g`} />
                <MacroPill label="Fat" value={`${meal.fat}g`} />
                <MacroPill label="Calories" value={`${meal.calories}`} />
              </div>
            </article>
          ))}
        </div>
      </SectionCard>

      <SectionCard title="Trends" subtitle="Select a metric and time range to view trend direction">
        <div className="space-y-3 rounded-xl border border-stone-200 bg-white p-3 sm:p-4">
          <div className="flex flex-wrap gap-2">
            {(Object.keys(metricLabels) as TrendMetric[]).map((metric) => (
              <button
                key={metric}
                type="button"
                onClick={() => setActiveMetric(metric)}
                className={`rounded-full px-3 py-1.5 text-xs font-medium transition sm:text-sm ${
                  activeMetric === metric
                    ? "bg-stone-900 text-white"
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
                    ? "border border-amber-300 bg-amber-100 text-amber-900"
                    : "border border-stone-300 bg-white text-stone-600 hover:bg-stone-50"
                }`}
              >
                {rangeLabels[range]}
              </button>
            ))}
          </div>

          <LineTrendChart data={trendData[activeRange]} metric={activeMetric} />
        </div>
      </SectionCard>
    </main>
  );
}

function LineTrendChart({
  data,
  metric
}: {
  data: Array<{ label: string; calories: number; protein: number; carbs: number; fat: number; weight: number }>;
  metric: TrendMetric;
}) {
  const values = data.map((item) => item[metric]);
  const min = Math.min(...values);
  const max = Math.max(...values);
  const spread = max - min || 1;
  const chartWidth = 100;
  const chartHeight = 50;

  const points = values
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
    <div className="rounded-xl border border-stone-200 bg-gradient-to-b from-stone-50 to-amber-50/30 p-3 sm:p-4">
      <div className="mb-2 flex items-end justify-between gap-3">
        <div>
          <p className="text-xs uppercase tracking-[0.1em] text-stone-500">{metricLabels[metric]}</p>
          <p className="text-xl font-semibold text-stone-900">
            {latestValue}
            <span className="ml-1 text-sm font-medium text-stone-500">{metric === "weight" ? "lb" : metric === "calories" ? "kcal" : "g"}</span>
          </p>
        </div>
        <p className={`text-xs font-medium ${delta >= 0 ? "text-amber-800" : "text-emerald-700"}`}>
          {deltaPrefix}
          {delta.toFixed(metric === "weight" ? 1 : 0)} vs prior
        </p>
      </div>

      <div className="relative h-44 w-full rounded-lg border border-stone-200 bg-white p-2">
        <svg viewBox="0 0 100 50" preserveAspectRatio="none" className="h-full w-full">
          <defs>
            <linearGradient id="trend-fill" x1="0" x2="0" y1="0" y2="1">
              <stop offset="0%" stopColor="#f59e0b" stopOpacity="0.26" />
              <stop offset="100%" stopColor="#f59e0b" stopOpacity="0.03" />
            </linearGradient>
          </defs>
          <polyline fill="none" stroke="#f59e0b" strokeWidth="1.5" points={points} />
          <polygon fill="url(#trend-fill)" points={`0,50 ${points} 100,50`} />
        </svg>
        <div className="mt-2 grid grid-cols-4 gap-1 text-[10px] text-stone-500 sm:grid-cols-6 md:grid-cols-8">
          {data.map((point) => (
            <span key={point.label} className="truncate text-center">
              {point.label}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}

function MacroPill({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-stone-200 bg-stone-50 px-2.5 py-2">
      <p className="text-[10px] uppercase tracking-[0.08em] text-stone-500">{label}</p>
      <p className="mt-1 text-sm font-semibold text-stone-800">{value}</p>
    </div>
  );
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
    <section className="rounded-3xl border border-stone-200/80 bg-stone-50/65 p-5 shadow-sm sm:p-6">
      <header className="mb-4">
        <h3 className="text-xl font-semibold tracking-tight text-stone-900">{title}</h3>
        <p className="mt-1 text-sm text-stone-600 sm:text-base">{subtitle}</p>
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
    <article className="rounded-xl border border-stone-200 bg-white p-3 shadow-sm">
      <p className="text-[11px] font-medium uppercase tracking-[0.1em] text-stone-500">{label}</p>
      <p className="mt-1 text-lg font-semibold text-stone-900">
        {consumed}
        <span className="ml-1 text-sm font-medium text-stone-500">/ {target}</span>
      </p>
      <p className="text-xs text-stone-600">{remaining} {unit} remaining</p>
      <div className="mt-3 h-2 rounded-full bg-stone-200">
        <div className="h-2 rounded-full bg-amber-300" style={{ width: `${pct}%` }} />
      </div>
      <p className="mt-1 text-[11px] text-stone-500">{pct}% complete</p>
    </article>
  );
}
