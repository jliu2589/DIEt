CREATE TABLE IF NOT EXISTS weight_entries (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    weight NUMERIC(10,2) NOT NULL,
    unit TEXT NOT NULL DEFAULT 'kg',
    logged_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_weight_entries_user_id ON weight_entries (user_id);
CREATE INDEX IF NOT EXISTS idx_weight_entries_logged_at ON weight_entries (logged_at DESC);
CREATE INDEX IF NOT EXISTS idx_weight_entries_user_logged_at ON weight_entries (user_id, logged_at DESC);

CREATE TRIGGER trg_weight_entries_updated_at
BEFORE UPDATE ON weight_entries
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
