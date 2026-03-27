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

export type DailySummaryResponse = NutritionTotals & {
  user_id: string;
  date: string;
};

export type CreateMealRequest = {
  user_id: string;
  source: string;
  raw_text: string;
  eaten_at: string;
};

export type CreateMealResponse = {
  meal_event_id: number;
  processed_from: "cache" | "openai";
  canonical_name: string;
  nutrition: NutritionTotals;
  confidence_score?: number;
};

export type RecentMeal = {
  meal_event_id: number;
  canonical_name: string;
  eaten_at: string;
  calories_kcal: number;
  protein_g: number;
  carbohydrate_g: number;
  fat_g: number;
  source: string;
};

export type RecentMealsResponse = {
  items: RecentMeal[];
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

  return response.json();
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

  return response.json();
}

export async function getRecentMeals(userId: string, limit = 20): Promise<RecentMealsResponse> {
  const params = new URLSearchParams({ user_id: userId, limit: String(limit) });
  const response = await fetch(buildUrl(`/v1/meals/recent?${params.toString()}`), {
    cache: "no-store"
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch recent meals (${response.status})`);
  }

  return response.json();
}
