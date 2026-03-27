CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS meal_events (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    source TEXT NOT NULL,
    source_message_id TEXT,
    event_type TEXT NOT NULL,
    raw_text TEXT,
    image_url TEXT,
    eaten_at TIMESTAMPTZ NOT NULL,
    processing_status TEXT NOT NULL DEFAULT 'pending',
    fingerprint_hash TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS meal_analysis (
    id BIGSERIAL PRIMARY KEY,
    meal_event_id BIGINT NOT NULL UNIQUE REFERENCES meal_events(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    canonical_name TEXT NOT NULL,
    confidence_score NUMERIC(4,3),
    assumptions_json JSONB,
    items_json JSONB,
    raw_analysis_json JSONB,
    calories_kcal NUMERIC(10,2),
    protein_g NUMERIC(10,2),
    carbohydrate_g NUMERIC(10,2),
    fat_g NUMERIC(10,2),
    fiber_g NUMERIC(10,2),
    sugars_g NUMERIC(10,2),
    saturated_fat_g NUMERIC(10,2),
    sodium_mg NUMERIC(10,2),
    potassium_mg NUMERIC(10,2),
    calcium_mg NUMERIC(10,2),
    magnesium_mg NUMERIC(10,2),
    iron_mg NUMERIC(10,2),
    zinc_mg NUMERIC(10,2),
    vitamin_d_mcg NUMERIC(10,2),
    vitamin_b12_mcg NUMERIC(10,2),
    folate_b9_mcg NUMERIC(10,2),
    vitamin_c_mg NUMERIC(10,2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS meal_memory (
    id BIGSERIAL PRIMARY KEY,
    fingerprint_hash TEXT NOT NULL UNIQUE,
    canonical_name TEXT NOT NULL,
    confidence_score NUMERIC(4,3),
    assumptions_json JSONB,
    items_json JSONB,
    raw_analysis_json JSONB,
    calories_kcal NUMERIC(10,2),
    protein_g NUMERIC(10,2),
    carbohydrate_g NUMERIC(10,2),
    fat_g NUMERIC(10,2),
    fiber_g NUMERIC(10,2),
    sugars_g NUMERIC(10,2),
    saturated_fat_g NUMERIC(10,2),
    sodium_mg NUMERIC(10,2),
    potassium_mg NUMERIC(10,2),
    calcium_mg NUMERIC(10,2),
    magnesium_mg NUMERIC(10,2),
    iron_mg NUMERIC(10,2),
    zinc_mg NUMERIC(10,2),
    vitamin_d_mcg NUMERIC(10,2),
    vitamin_b12_mcg NUMERIC(10,2),
    folate_b9_mcg NUMERIC(10,2),
    vitamin_c_mg NUMERIC(10,2),
    usage_count INT NOT NULL DEFAULT 1,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS daily_nutrition_summary (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    summary_date DATE NOT NULL,
    calories_kcal NUMERIC(10,2) DEFAULT 0,
    protein_g NUMERIC(10,2) DEFAULT 0,
    carbohydrate_g NUMERIC(10,2) DEFAULT 0,
    fat_g NUMERIC(10,2) DEFAULT 0,
    fiber_g NUMERIC(10,2) DEFAULT 0,
    sugars_g NUMERIC(10,2) DEFAULT 0,
    saturated_fat_g NUMERIC(10,2) DEFAULT 0,
    sodium_mg NUMERIC(10,2) DEFAULT 0,
    potassium_mg NUMERIC(10,2) DEFAULT 0,
    calcium_mg NUMERIC(10,2) DEFAULT 0,
    magnesium_mg NUMERIC(10,2) DEFAULT 0,
    iron_mg NUMERIC(10,2) DEFAULT 0,
    zinc_mg NUMERIC(10,2) DEFAULT 0,
    vitamin_d_mcg NUMERIC(10,2) DEFAULT 0,
    vitamin_b12_mcg NUMERIC(10,2) DEFAULT 0,
    folate_b9_mcg NUMERIC(10,2) DEFAULT 0,
    vitamin_c_mg NUMERIC(10,2) DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, summary_date)
);

CREATE INDEX IF NOT EXISTS idx_meal_events_user_eaten_at ON meal_events (user_id, eaten_at DESC);
CREATE INDEX IF NOT EXISTS idx_meal_events_processing_status ON meal_events (processing_status);
CREATE INDEX IF NOT EXISTS idx_meal_events_fingerprint_hash ON meal_events (fingerprint_hash);
CREATE INDEX IF NOT EXISTS idx_meal_analysis_user_created_at ON meal_analysis (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_meal_memory_last_used_at ON meal_memory (last_used_at DESC);
CREATE INDEX IF NOT EXISTS idx_daily_summary_user_date ON daily_nutrition_summary (user_id, summary_date DESC);

CREATE TRIGGER trg_meal_events_updated_at
BEFORE UPDATE ON meal_events
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_meal_analysis_updated_at
BEFORE UPDATE ON meal_analysis
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_meal_memory_updated_at
BEFORE UPDATE ON meal_memory
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_daily_summary_updated_at
BEFORE UPDATE ON daily_nutrition_summary
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
