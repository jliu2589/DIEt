package repositories

import "github.com/jackc/pgx/v5/pgxpool"

type Repositories struct {
	MealEvents            *MealEventsRepository
	MealAnalysis          *MealAnalysisRepository
	MealMemory            *MealMemoryRepository
	DailyNutritionSummary *DailyNutritionSummaryRepository
	UserSettings          *UserSettingsRepository
	WeightEntries         *WeightEntriesRepository
	CanonicalFoods        *CanonicalFoodsRepository
	Meals                 *MealsRepository
	MealItems             *MealItemsRepository
}

func New(pool *pgxpool.Pool) Repositories {
	return Repositories{
		MealEvents:            NewMealEventsRepository(pool),
		MealAnalysis:          NewMealAnalysisRepository(pool),
		MealMemory:            NewMealMemoryRepository(pool),
		DailyNutritionSummary: NewDailyNutritionSummaryRepository(pool),
		UserSettings:          NewUserSettingsRepository(pool),
		WeightEntries:         NewWeightEntriesRepository(pool),
		CanonicalFoods:        NewCanonicalFoodsRepository(pool),
		Meals:                 NewMealsRepository(pool),
		MealItems:             NewMealItemsRepository(pool),
	}
}
