ALTER TABLE items
    ADD COLUMN short_description TEXT NOT NULL DEFAULT '';

UPDATE items
SET short_description = description
WHERE description <> '';
