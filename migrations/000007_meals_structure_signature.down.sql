DROP INDEX IF EXISTS idx_meals_structure_signature_unique;

ALTER TABLE meals
    DROP COLUMN IF EXISTS structure_signature;
