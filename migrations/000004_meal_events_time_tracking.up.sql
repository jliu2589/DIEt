ALTER TABLE meal_events
    ADD COLUMN IF NOT EXISTS logged_at TIMESTAMPTZ;

UPDATE meal_events
SET logged_at = COALESCE(logged_at, created_at, NOW())
WHERE logged_at IS NULL;

ALTER TABLE meal_events
    ALTER COLUMN logged_at SET NOT NULL,
    ALTER COLUMN logged_at SET DEFAULT NOW();

ALTER TABLE meal_events
    ADD COLUMN IF NOT EXISTS time_source TEXT NOT NULL DEFAULT 'default_now';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'meal_events_time_source_check'
    ) THEN
        ALTER TABLE meal_events
            ADD CONSTRAINT meal_events_time_source_check
            CHECK (time_source IN ('explicit', 'inferred', 'default_now', 'edited'));
    END IF;
END $$;

ALTER TABLE meal_events
    ADD COLUMN IF NOT EXISTS eaten_at TIMESTAMPTZ;

UPDATE meal_events
SET eaten_at = COALESCE(eaten_at, logged_at, created_at, NOW())
WHERE eaten_at IS NULL;

ALTER TABLE meal_events
    ALTER COLUMN eaten_at SET NOT NULL,
    ALTER COLUMN eaten_at SET DEFAULT NOW();
