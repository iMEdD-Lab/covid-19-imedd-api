CREATE TABLE IF NOT EXISTS deaths_per_municipality_cum
(
    date         DATE,
    municipality VARCHAR(255),
    deaths_cum   INTEGER,
    UNIQUE (date, municipality)
);

CREATE INDEX IF NOT EXISTS idx_date_deaths ON deaths_per_municipality_cum (date);
CREATE INDEX IF NOT EXISTS idx_municipality ON deaths_per_municipality_cum (municipality);