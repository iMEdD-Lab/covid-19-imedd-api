CREATE TABLE IF NOT EXISTS cases_per_prefecture
(
    geo_id INTEGER,
    date   DATE,
    cases  INTEGER,
    UNIQUE (geo_id, date)
);

CREATE INDEX IF NOT EXISTS idx_geo_id ON cases_per_prefecture (geo_id);
CREATE INDEX IF NOT EXISTS idx_date ON cases_per_prefecture (date);

ALTER TABLE cases_per_prefecture
    ADD CONSTRAINT fk_geo_id FOREIGN KEY (geo_id) REFERENCES counties (id);