ALTER TABLE meal_events
    DROP CONSTRAINT IF EXISTS meal_events_time_source_check;

ALTER TABLE meal_events
    DROP COLUMN IF EXISTS time_source;

ALTER TABLE meal_events
    DROP COLUMN IF EXISTS logged_at;
