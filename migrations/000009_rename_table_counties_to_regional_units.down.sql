ALTER TABLE IF EXISTS regional_units
    RENAME TO counties;

ALTER TABLE IF EXISTS counties
    RENAME COLUMN regional_unit to county;

ALTER TABLE IF EXISTS counties
    RENAME COLUMN regional_unit_normalized to county_normalized;

ALTER TABLE IF EXISTS cases_per_regional_unit
    RENAME COLUMN regional_unit_id TO county_id;

ALTER TABLE IF EXISTS cases_per_regional_unit
    RENAME TO cases_per_county;