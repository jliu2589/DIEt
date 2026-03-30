export type NutritionTotals = {
  calories_kcal: number;
  protein_g: number;
  carbohydrate_g: number;
  fat_g: number;
  fiber_g: number;
  sugars_g: number;
  saturated_fat_g: number;
  sodium_mg: number;
  potassium_mg: number;
  calcium_mg: number;
  magnesium_mg: number;
  iron_mg: number;
  zinc_mg: number;
  vitamin_d_mcg: number;
  vitamin_b12_mcg: number;
  folate_b9_mcg: number;
  vitamin_c_mg: number;
};

export type DailySummaryResponse = {
  user_id: string;
  date: string;
  totals: NutritionTotals;
  meal_count?: number;
} & NutritionTotals;

export type CreateMealRequest = {
  user_id: string;
  source: string;
  raw_text: string;
  eaten_at: string;
};

export type MealItemView = {
  meal_event_id: number;
  canonical_name: string;
  logged_at: string;
  eaten_at: string;
  time_source?: string | null;
  source: string;
  confidence_score?: number;
  calories_kcal: number | null;
  protein_g: number | null;
  carbohydrate_g: number | null;
  fat_g: number | null;
};

export type CreateMealResponse = {
  intent: string;
  logged: boolean;
  message: string;
  item?: MealItemView;
};

export type RecentMeal = MealItemView;

export type RecentMealsResponse = {
  items: RecentMeal[];
};

export type DashboardTodayResponse = {
  user_id: string;
  date: string;
  daily_summary: DailySummaryResponse;
  recent_meals: RecentMeal[];
};

export type ChatResponse = {
  intent: string;
  message_to_user: string;
  meal_result?: {
    meal_event_id: number;
    canonical_name: string;
    calories_kcal?: number;
    protein_g?: number;
    carbohydrate_g?: number;
    fat_g?: number;
  };
  weight_result?: {
    weight: number;
    unit: string;
    logged_at: string;
  };
  recommendation_result?: {
    text: string;
  };
};

export type TrendsPoint = {
  date: string;
  weight: number | null;
  calories_kcal: number | null;
  protein_g: number | null;
  carbohydrate_g: number | null;
  fat_g: number | null;
};

export type TrendsResponse = {
  user_id: string;
  range: "7d" | "30d" | "90d" | "1y";
  points: TrendsPoint[];
};

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL;

function buildUrl(path: string) {
  if (!API_BASE_URL) {
    throw new Error("NEXT_PUBLIC_API_BASE_URL is not set");
  }
  return `${API_BASE_URL}${path}`;
}

export async function getDailySummary(userId: string, date: string): Promise<DailySummaryResponse> {
  const params = new URLSearchParams({ user_id: userId, date });
  const response = await fetch(buildUrl(`/v1/daily-summary?${params.toString()}`), {
    cache: "no-store"
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch daily summary (${response.status})`);
  }

  return response.json() as Promise<DailySummaryResponse>;
}

export async function createMeal(payload: CreateMealRequest): Promise<CreateMealResponse> {
  const response = await fetch(buildUrl("/v1/meals"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload)
  });

  if (!response.ok) {
    throw new Error(`Failed to create meal (${response.status})`);
  }

  return response.json() as Promise<CreateMealResponse>;
}

export async function getRecentMeals(userId: string, limit = 20): Promise<RecentMealsResponse> {
  const params = new URLSearchParams({ user_id: userId, limit: String(limit) });
  const response = await fetch(buildUrl(`/v1/meals/recent?${params.toString()}`), {
    cache: "no-store"
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch recent meals (${response.status})`);
  }

  return response.json() as Promise<RecentMealsResponse>;
}

export async function getDashboardToday(userId: string): Promise<DashboardTodayResponse> {
  const params = new URLSearchParams({ user_id: userId });
  const response = await fetch(buildUrl(`/v1/dashboard/today?${params.toString()}`), {
    cache: "no-store"
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch dashboard (${response.status})`);
  }

  return response.json() as Promise<DashboardTodayResponse>;
}

export async function postChat(userId: string, message: string): Promise<ChatResponse> {
  const response = await fetch(buildUrl("/v1/chat"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ user_id: userId, message })
  });

  if (!response.ok) {
    throw new Error(`Failed to send chat (${response.status})`);
  }

  return response.json() as Promise<ChatResponse>;
}

export async function getTrends(userId: string, range: "7d" | "30d" | "90d" | "1y"): Promise<TrendsResponse> {
  const params = new URLSearchParams({ user_id: userId, range });
  const response = await fetch(buildUrl(`/v1/trends?${params.toString()}`), {
    cache: "no-store"
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch trends (${response.status})`);
  }

  return response.json() as Promise<TrendsResponse>;
}
