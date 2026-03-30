ALTER TABLE meals
    ADD COLUMN IF NOT EXISTS structure_signature TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_meals_structure_signature_unique
    ON meals (structure_signature)
    WHERE structure_signature IS NOT NULL;
