CREATE TABLE IF NOT EXISTS demography_per_age
(
    date                DATE NOT NULL,
    category            VARCHAR(5),
    cases               INTEGER,
    deaths              INTEGER,
    intensive           INTEGER,
    discharged          INTEGER,
    hospitalized        INTEGER,
    hospitalized_in_icu INTEGER,
    passed_away         INTEGER,
    recovered           INTEGER,
    treated_at_home     INTEGER,
    UNIQUE (date, category)
);

CREATE INDEX IF NOT EXISTS idx_demography_per_age_date ON demography_per_age (date);
CREATE INDEX IF NOT EXISTS idx_demography_per_age_category ON demography_per_age (category);