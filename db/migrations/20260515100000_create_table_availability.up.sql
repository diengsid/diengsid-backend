CREATE TABLE availabilities (
    id              VARCHAR(36) PRIMARY KEY,
    rentable_id     VARCHAR(36) NOT NULL,
    date            BIGINT NOT NULL,
    available_count INT NOT NULL,
    price_override  DOUBLE PRECISION,
    created_at      BIGINT,
    updated_at      BIGINT,

    CONSTRAINT fk_availabilities_rentable
        FOREIGN KEY (rentable_id)
        REFERENCES rentables(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_availability_rentable_date
        UNIQUE (rentable_id, date)
);
