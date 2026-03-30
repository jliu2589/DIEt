CREATE TABLE IF NOT EXISTS user_settings (
    user_id TEXT PRIMARY KEY,
    name TEXT,
    height_cm NUMERIC(10,2),
    weight_goal_kg NUMERIC(10,2),
    calorie_goal NUMERIC(10,2),
    protein_goal_g NUMERIC(10,2),
    weight_unit TEXT NOT NULL DEFAULT 'kg',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_user_settings_updated_at
BEFORE UPDATE ON user_settings
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
