DROP TRIGGER IF EXISTS trg_daily_summary_updated_at ON daily_nutrition_summary;
DROP TRIGGER IF EXISTS trg_meal_memory_updated_at ON meal_memory;
DROP TRIGGER IF EXISTS trg_meal_analysis_updated_at ON meal_analysis;
DROP TRIGGER IF EXISTS trg_meal_events_updated_at ON meal_events;

DROP TABLE IF EXISTS daily_nutrition_summary;
DROP TABLE IF EXISTS meal_memory;
DROP TABLE IF EXISTS meal_analysis;
DROP TABLE IF EXISTS meal_events;

DROP FUNCTION IF EXISTS set_updated_at;
