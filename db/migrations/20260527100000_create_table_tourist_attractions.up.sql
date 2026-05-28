CREATE TABLE tourist_attractions (
    id          VARCHAR(36) PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    address     VARCHAR(255),
    latitude    DOUBLE PRECISION,
    longitude   DOUBLE PRECISION,
    category    VARCHAR(100),
    image_url   VARCHAR(255),
    created_at  BIGINT NOT NULL,
    updated_at  BIGINT NOT NULL
);

CREATE TABLE property_nearby_attractions (
    property_id           VARCHAR(36) NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    tourist_attraction_id VARCHAR(36) NOT NULL REFERENCES tourist_attractions(id) ON DELETE CASCADE,
    distance_km           DOUBLE PRECISION,
    duration_minutes      INT,
    sort_order            INT NOT NULL DEFAULT 0,
    PRIMARY KEY (property_id, tourist_attraction_id)
);
