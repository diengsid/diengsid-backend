ALTER TABLE properties ADD COLUMN IF NOT EXISTS slug VARCHAR(255);

UPDATE properties SET slug = id WHERE slug IS NULL;

ALTER TABLE properties ALTER COLUMN slug SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_properties_slug ON properties (slug);
