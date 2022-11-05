CREATE TABLE IF NOT EXISTS cases_per_county
(
    county_id INTEGER,
    date   DATE,
    cases  INTEGER,
    UNIQUE (county_id, date)
);

CREATE INDEX IF NOT EXISTS idx_county_id ON cases_per_county (county_id);
CREATE INDEX IF NOT EXISTS idx_date ON cases_per_county (date);

ALTER TABLE cases_per_county
    ADD CONSTRAINT fk_county_id FOREIGN KEY (county_id) REFERENCES counties (id);