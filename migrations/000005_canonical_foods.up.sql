CREATE TABLE IF NOT EXISTS foods (
    id BIGSERIAL PRIMARY KEY,
    canonical_name TEXT NOT NULL,
    default_amount NUMERIC(10,2) NOT NULL,
    default_unit TEXT NOT NULL,
    category TEXT,
    source_type TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS food_nutrition (
    food_id BIGINT PRIMARY KEY REFERENCES foods(id) ON DELETE CASCADE,
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

CREATE INDEX IF NOT EXISTS idx_foods_canonical_name ON foods (canonical_name);

CREATE TRIGGER trg_foods_updated_at
BEFORE UPDATE ON foods
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_food_nutrition_updated_at
BEFORE UPDATE ON food_nutrition
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
