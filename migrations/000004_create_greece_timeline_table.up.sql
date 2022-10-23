CREATE TABLE IF NOT EXISTS greece_timeline
(
    date                      DATE PRIMARY KEY,
    cases                     INTEGER,
    total_reinfections        INTEGER,
    deaths                    INTEGER,
    deaths_cum                INTEGER,
    recovered                 INTEGER,
    beds_occupancy            FLOAT,
    icu_occupancy             FLOAT,
    intubated                 INTEGER,
    intubated_vac             INTEGER,
    intubated_unvac           INTEGER,
    hospital_admissions       INTEGER,
    hospital_discharges       INTEGER,
    estimated_new_rtpcr_tests INTEGER,
    estimated_new_rapid_tests INTEGER,
    estimated_new_total_tests INTEGER,
    UNIQUE (date)
);