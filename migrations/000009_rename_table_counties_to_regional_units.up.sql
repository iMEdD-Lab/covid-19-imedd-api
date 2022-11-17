ALTER TABLE IF EXISTS counties
    RENAME COLUMN county to regional_unit;

ALTER TABLE IF EXISTS counties
    RENAME COLUMN county_normalized to regional_unit_normalized;

ALTER TABLE IF EXISTS counties
    RENAME TO regional_units;

ALTER TABLE IF EXISTS cases_per_county
    RENAME TO cases_per_regional_unit;

ALTER TABLE IF EXISTS cases_per_regional_unit
    RENAME COLUMN county_id TO regional_unit_id;