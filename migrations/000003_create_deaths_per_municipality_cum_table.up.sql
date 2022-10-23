CREATE TABLE IF NOT EXISTS municipalities
(
    id   SERIAL PRIMARY KEY,
    name VARCHAR(256),
    slug VARCHAR(256),
    UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS deaths_per_municipality_cum
(
    year            INTEGER,
    municipality_id INTEGER,
    deaths_cum      INTEGER,
    UNIQUE (year, municipality_id)
);

CREATE INDEX IF NOT EXISTS idx_deaths_per_municipality_cum_year ON deaths_per_municipality_cum (year);
CREATE INDEX IF NOT EXISTS idx_deaths_per_municipality_cum_municipality_id ON deaths_per_municipality_cum (municipality_id);

ALTER TABLE deaths_per_municipality_cum
    ADD CONSTRAINT fk_municipality_id FOREIGN KEY (municipality_id) REFERENCES municipalities (id);