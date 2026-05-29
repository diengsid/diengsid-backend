-- Remove FK constraint and experience_id from properties
ALTER TABLE properties DROP CONSTRAINT IF EXISTS fk_properties_experience;
ALTER TABLE properties DROP COLUMN IF EXISTS experience_id;

-- Add new fields to properties
ALTER TABLE properties
    ADD COLUMN IF NOT EXISTS title VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS address VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS thumbnail_url VARCHAR(255),
    ADD COLUMN IF NOT EXISTS lat DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS lng DOUBLE PRECISION;

-- Remove DEFAULT after adding (cleanup)
ALTER TABLE properties
    ALTER COLUMN title DROP DEFAULT,
    ALTER COLUMN address DROP DEFAULT,
    ALTER COLUMN description DROP DEFAULT;

-- Create property_images table
CREATE TABLE IF NOT EXISTS property_images (
    id VARCHAR(36) PRIMARY KEY,
    property_id VARCHAR(36) NOT NULL,
    image_url VARCHAR(255) NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,

    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,

    CONSTRAINT fk_property_images_property
        FOREIGN KEY (property_id)
        REFERENCES properties(id)
        ON DELETE CASCADE
);

-- Drop experience tables
DROP TABLE IF EXISTS experience_images;
DROP TABLE IF EXISTS experiences;
