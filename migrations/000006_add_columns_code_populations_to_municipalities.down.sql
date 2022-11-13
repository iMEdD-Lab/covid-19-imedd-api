DROP INDEX IF EXISTS idx_municipalities_code;

ALTER TABLE municipalities
    DROP COLUMN IF EXISTS code,
    DROP COLUMN IF EXISTS pop_11,
    DROP COLUMN IF EXISTS pop_12;
