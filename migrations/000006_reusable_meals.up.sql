CREATE TABLE IF NOT EXISTS meals (
    id BIGSERIAL PRIMARY KEY,
    canonical_name TEXT NOT NULL,
    fingerprint_hash TEXT UNIQUE,
    source_type TEXT,
    confidence_score NUMERIC(4,3),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS meal_items (
    id BIGSERIAL PRIMARY KEY,
    meal_id BIGINT NOT NULL REFERENCES meals(id) ON DELETE CASCADE,
    food_id BIGINT NOT NULL REFERENCES foods(id) ON DELETE CASCADE,
    quantity NUMERIC(10,2) NOT NULL,
    unit TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_meal_items_meal_id ON meal_items (meal_id);
CREATE INDEX IF NOT EXISTS idx_meal_items_food_id ON meal_items (food_id);

CREATE TRIGGER trg_meals_updated_at
BEFORE UPDATE ON meals
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
