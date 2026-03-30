const goals = [
  { label: "Calories", consumed: 1380, target: 2100, unit: "kcal" },
  { label: "Protein", consumed: 78, target: 130, unit: "g" },
  { label: "Carbs", consumed: 145, target: 220, unit: "g" },
  { label: "Fat", consumed: 52, target: 75, unit: "g" }
];

const chatMessages = [
  { id: 1, from: "assistant", text: "Good morning — you’re on track. Want a high-protein lunch idea?" },
  { id: 2, from: "user", text: "Yes, something quick." },
  { id: 3, from: "assistant", text: "Try grilled chicken, quinoa, and roasted vegetables (~520 kcal, 45g protein)." }
] as const;

const todaysMeals = [
  { id: 1, name: "Greek Yogurt + Berries", time: "8:10 AM", calories: 320, protein: 28 },
  { id: 2, name: "Salmon Rice Bowl", time: "12:42 PM", calories: 610, protein: 41 },
  { id: 3, name: "Almonds + Banana", time: "3:20 PM", calories: 210, protein: 9 }
];

const weeklyTrend = [68, 72, 65, 74, 70, 78, 76];

export default function HomePage() {
  return (
    <main className="space-y-5 pb-8">
      <section className="rounded-2xl border border-stone-200/80 bg-stone-50/70 p-4 shadow-sm sm:p-6">
        <p className="text-xs font-medium uppercase tracking-[0.14em] text-stone-500">Today’s Progress</p>
        <h2 className="mt-2 text-2xl font-semibold tracking-tight text-stone-900">Nutrition Dashboard</h2>
        <div className="mt-5 grid grid-cols-2 gap-3 sm:grid-cols-4">
          {goals.map((goal) => (
            <GoalCard key={goal.label} {...goal} />
          ))}
        </div>
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
              className="rounded-xl border border-stone-200 bg-white px-3 py-3 text-sm sm:flex sm:items-center sm:justify-between"
            >
              <div>
                <p className="font-medium text-stone-900">{meal.name}</p>
                <p className="text-xs text-stone-500">{meal.time}</p>
              </div>
              <div className="mt-2 flex gap-3 text-xs text-stone-600 sm:mt-0">
                <span>{meal.calories} kcal</span>
                <span>{meal.protein}g protein</span>
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

  return (
    <article className="rounded-xl border border-stone-200 bg-white p-3">
      <p className="text-xs text-stone-500">{label}</p>
      <p className="mt-1 text-base font-semibold text-stone-900">
        {consumed}
        <span className="ml-1 text-xs font-medium text-stone-500">/ {target + " " + unit}</span>
      </p>
      <div className="mt-2 h-2 rounded-full bg-stone-200">
        <div className="h-2 rounded-full bg-amber-300" style={{ width: `${pct}%` }} />
      </div>
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
