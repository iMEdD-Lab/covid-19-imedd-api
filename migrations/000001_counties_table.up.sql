CREATE TABLE IF NOT EXISTS counties
(
    id                SERIAL PRIMARY KEY,
    slug              VARCHAR(255) NOT NULL UNIQUE,
    department        VARCHAR(255) NOT NULL,
    prefecture        VARCHAR(255) NOT NULL,
    county_normalized VARCHAR(255) NOT NULL UNIQUE,
    county            VARCHAR(255) NOT NULL,
    pop_11            INTEGER      NOT NULL,
    UNIQUE (county_normalized)
);

CREATE INDEX IF NOT EXISTS idx_slug ON counties (slug);