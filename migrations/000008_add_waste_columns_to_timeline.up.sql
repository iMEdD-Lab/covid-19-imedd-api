ALTER TABLE greece_timeline
    ADD COLUMN IF NOT EXISTS waste_highest_place VARCHAR(100);
ALTER TABLE greece_timeline
    ADD COLUMN IF NOT EXISTS waste_highest_percentage FLOAT;