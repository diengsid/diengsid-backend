CREATE TABLE amenities (
    id          VARCHAR(36) PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    icon        VARCHAR(100),
    category    VARCHAR(50),
    created_at  BIGINT NOT NULL,
    updated_at  BIGINT NOT NULL
);

CREATE TABLE property_amenities (
    property_id VARCHAR(36) NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    amenity_id  VARCHAR(36) NOT NULL REFERENCES amenities(id) ON DELETE CASCADE,
    PRIMARY KEY (property_id, amenity_id)
);

CREATE TABLE rentable_amenities (
    rentable_id VARCHAR(36) NOT NULL REFERENCES rentables(id) ON DELETE CASCADE,
    amenity_id  VARCHAR(36) NOT NULL REFERENCES amenities(id) ON DELETE CASCADE,
    PRIMARY KEY (rentable_id, amenity_id)
);
