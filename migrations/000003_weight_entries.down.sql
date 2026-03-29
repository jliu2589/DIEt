DROP TRIGGER IF EXISTS trg_weight_entries_updated_at ON weight_entries;
DROP INDEX IF EXISTS idx_weight_entries_user_logged_at;
DROP INDEX IF EXISTS idx_weight_entries_logged_at;
DROP INDEX IF EXISTS idx_weight_entries_user_id;
DROP TABLE IF EXISTS weight_entries;
