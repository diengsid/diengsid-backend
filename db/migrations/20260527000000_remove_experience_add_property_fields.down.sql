-- Restore experience tables
CREATE TABLE IF NOT EXISTS experiences (
    id VARCHAR(36) PRIMARY KEY,
    experience_type VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    thumbnail_url VARCHAR(255),
    lat DOUBLE PRECISION,
    lng DOUBLE PRECISION,
    base_price DOUBLE PRECISION NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS experience_images (
    id VARCHAR(36) PRIMARY KEY,
    experience_id VARCHAR(36) NOT NULL,
    image_url VARCHAR(255) NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    CONSTRAINT fk_experience_images_experience
        FOREIGN KEY (experience_id)
        REFERENCES experiences(id)
        ON DELETE CASCADE
);

-- Drop property_images
DROP TABLE IF EXISTS property_images;

-- Remove added columns from properties
ALTER TABLE properties
    DROP COLUMN IF EXISTS title,
    DROP COLUMN IF EXISTS address,
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS thumbnail_url,
    DROP COLUMN IF EXISTS lat,
    DROP COLUMN IF EXISTS lng;

-- Restore experience_id to properties
ALTER TABLE properties ADD COLUMN IF NOT EXISTS experience_id VARCHAR(36) NOT NULL DEFAULT '';
ALTER TABLE properties ADD CONSTRAINT fk_properties_experience
    FOREIGN KEY (experience_id) REFERENCES experiences(id) ON DELETE CASCADE;
