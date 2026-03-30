"use client";

import { FormEvent, useState } from "react";

const goals = [
  { label: "Calories", consumed: 1380, target: 2500, unit: "kcal" },
  { label: "Protein", consumed: 92, target: 160, unit: "g" },
  { label: "Carbs", consumed: 145, target: 220, unit: "g" },
  { label: "Fat", consumed: 52, target: 75, unit: "g" }
];

const chatMessages = [
  { id: 1, from: "assistant", text: "Good morning — you’re on track. Want a high-protein lunch idea?" },
  { id: 2, from: "user", text: "Yes, something quick." },
  { id: 3, from: "assistant", text: "Try grilled chicken, quinoa, and roasted vegetables (~520 kcal, 45g protein)." }
] as const;

const todaysMeals = [
  { id: 1, name: "Greek Yogurt + Berries", time: "8:10 AM", calories: 320, protein: 28, carbs: 32, fat: 9 },
  { id: 2, name: "Salmon Rice Bowl", time: "12:42 PM", calories: 610, protein: 41, carbs: 58, fat: 23 },
  { id: 3, name: "Almonds + Banana", time: "3:20 PM", calories: 210, protein: 9, carbs: 24, fat: 11 },
  { id: 4, name: "Chicken & Avocado Wrap", time: "7:05 PM", calories: 470, protein: 36, carbs: 34, fat: 21 }
];

const weeklyTrend = [68, 72, 65, 74, 70, 78, 76];

export default function HomePage() {
  const [chatInput, setChatInput] = useState("");
  const [drafts, setDrafts] = useState<string[]>([]);

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
    <main className="space-y-5 pb-8">
      <section className="rounded-2xl border border-stone-200/80 bg-gradient-to-b from-stone-50 to-amber-50/40 p-4 shadow-sm sm:p-6">
        <p className="text-xs font-medium uppercase tracking-[0.14em] text-stone-500">Today’s Goals & Totals</p>
        <h2 className="mt-2 text-2xl font-semibold tracking-tight text-stone-900">Nutrition Dashboard</h2>
        <p className="mt-1 text-sm text-stone-600">Quick progress snapshot against your daily targets.</p>
        <div className="mt-5 grid grid-cols-2 gap-3 lg:grid-cols-4">
          {goals.map((goal) => (
            <GoalCard key={goal.label} {...goal} />
          ))}
        </div>
      </section>

      <section className="rounded-2xl border border-amber-200/70 bg-gradient-to-b from-white to-amber-50/70 p-4 shadow-sm sm:p-6">
        <div className="mb-3 space-y-1">
          <p className="text-xs font-medium uppercase tracking-[0.14em] text-amber-700/80">Main Chat</p>
          <h3 className="text-xl font-semibold tracking-tight text-stone-900 sm:text-2xl">Tell me what you need</h3>
          <p className="text-sm text-stone-600">
            Log meals, track your weight, ask for recommendations, or just chat about your nutrition day.
          </p>
        </div>

        <form onSubmit={onSubmitChat} className="space-y-3">
          <label htmlFor="chat-input" className="sr-only">
            Main nutrition chat input
          </label>
          <div className="rounded-2xl border border-stone-300 bg-white p-2 shadow-sm focus-within:border-amber-400">
            <textarea
              id="chat-input"
              value={chatInput}
              onChange={(event) => setChatInput(event.target.value)}
              rows={3}
              placeholder="Log a meal, your weight, or ask what to eat next"
              className="w-full resize-none rounded-xl border-none bg-transparent px-2 py-1 text-sm text-stone-800 outline-none placeholder:text-stone-400 sm:text-base"
            />
            <div className="flex items-center justify-between gap-3 px-2 pb-1 pt-2">
              <p className="text-xs text-stone-500">Try: “Lunch was chicken bowl” or “I weigh 176.2 lb”</p>
              <button
                type="submit"
                className="shrink-0 rounded-xl bg-stone-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-stone-700"
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

      <SectionCard title="Coach Chat" subtitle="Placeholder conversation UI (no backend wiring yet)">
        <div className="space-y-3">
          <div className="space-y-2 rounded-xl border border-stone-200 bg-white/70 p-3">
            {chatMessages.map((message) => (
              <ChatBubble key={message.id} from={message.from} text={message.text} />
            ))}
          </div>
          <div className="flex items-center gap-2">
            <input
              disabled
              placeholder="Type a message…"
              className="w-full rounded-xl border border-stone-300 bg-stone-100 px-3 py-2 text-sm text-stone-600 placeholder:text-stone-400"
            />
            <button
              disabled
              className="rounded-xl border border-stone-300 bg-stone-100 px-4 py-2 text-sm font-medium text-stone-500"
            >
              Send
            </button>
          </div>
        </div>
      </SectionCard>

      <SectionCard title="Today’s Meals" subtitle="Snapshot list with placeholder totals">
        <div className="space-y-2">
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

      <SectionCard title="Trends" subtitle="Weekly consistency chart placeholder">
        <div className="rounded-xl border border-stone-200 bg-white p-4">
          <div className="flex h-40 items-end gap-2">
            {weeklyTrend.map((value, idx) => (
              <div key={idx} className="flex flex-1 flex-col items-center justify-end gap-2">
                <div
                  className="w-full rounded-md bg-gradient-to-t from-amber-300 to-orange-200"
                  style={{ height: `${value}%` }}
                />
                <span className="text-[10px] text-stone-500">D{idx + 1}</span>
              </div>
            ))}
          </div>
        </div>
      </SectionCard>
    </main>
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
    <section className="rounded-2xl border border-stone-200/80 bg-stone-50/60 p-4 shadow-sm sm:p-5">
      <header className="mb-3">
        <h3 className="text-lg font-semibold text-stone-900">{title}</h3>
        <p className="text-sm text-stone-600">{subtitle}</p>
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

function ChatBubble({ from, text }: { from: "assistant" | "user"; text: string }) {
  const isUser = from === "user";
  return (
    <div className={`flex ${isUser ? "justify-end" : "justify-start"}`}>
      <p
        className={`max-w-[88%] rounded-2xl px-3 py-2 text-sm leading-relaxed ${
          isUser ? "bg-stone-900 text-stone-50" : "border border-stone-200 bg-stone-50 text-stone-800"
        }`}
      >
        {text}
      </p>
    </div>
  );
}
