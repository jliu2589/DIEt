CREATE TABLE IF NOT EXISTS idempotency_keys (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    request_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'processing',
    http_status INT,
    response_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, endpoint, idempotency_key)
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'idempotency_keys_status_check'
    ) THEN
        ALTER TABLE idempotency_keys
            ADD CONSTRAINT idempotency_keys_status_check
            CHECK (status IN ('processing', 'succeeded'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_idempotency_keys_user_created_at
    ON idempotency_keys (user_id, created_at DESC);

CREATE TRIGGER trg_idempotency_keys_updated_at
BEFORE UPDATE ON idempotency_keys
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
