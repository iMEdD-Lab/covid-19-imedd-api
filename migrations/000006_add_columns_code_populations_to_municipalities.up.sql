ALTER TABLE municipalities
    ADD COLUMN IF NOT EXISTS code   VARCHAR(30) UNIQUE,
    ADD COLUMN IF NOT EXISTS pop_11 INTEGER,
    ADD COLUMN IF NOT EXISTS pop_21 INTEGER;

CREATE INDEX IF NOT EXISTS idx_municipalities_code
    ON municipalities (code);
