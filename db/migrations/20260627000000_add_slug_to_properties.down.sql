DROP INDEX IF EXISTS idx_properties_slug;

ALTER TABLE properties DROP COLUMN IF EXISTS slug;
