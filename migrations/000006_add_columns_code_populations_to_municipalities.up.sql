ALTER TABLE municipalities
    ADD COLUMN IF NOT EXISTS code   VARCHAR(30),
    ADD COLUMN IF NOT EXISTS pop_11 INTEGER,
    ADD COLUMN IF NOT EXISTS pop_21 INTEGER;

CREATE INDEX IF NOT EXISTS idx_municipalities_code
    ON municipalities (code);

ALTER TABLE municipalities
    DROP CONSTRAINT IF EXISTS municipalities_name_key;