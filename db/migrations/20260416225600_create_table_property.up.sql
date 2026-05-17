CREATE TABLE host_profiles (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255),
    phone_number VARCHAR(255),
    profile_picture_url VARCHAR(255),
    address VARCHAR(255),
    bank_account_name VARCHAR(255),
    bank_account_number VARCHAR(255),
    ktp_number VARCHAR(255),
    bio VARCHAR(255),
    created_at BIGINT,
    updated_at BIGINT
);

CREATE TABLE properties (
    id VARCHAR(36) PRIMARY KEY,
    host_id VARCHAR(36) NOT NULL,
    experience_id VARCHAR(36) NOT NULL,
    property_type VARCHAR(20) DEFAULT 'homestay', -- ('hotel','villa','guesthouse','apartment','cabin','homestay')
    booking_type VARCHAR(20), -- 🔥 kunci utama : (room, unit)
    created_at BIGINT,
    updated_at BIGINT,

    CONSTRAINT fk_properties_host
        FOREIGN KEY (host_id)
        REFERENCES host_profiles(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_properties_experience
        FOREIGN KEY (experience_id)
        REFERENCES experiences(id)
        ON DELETE CASCADE,

    -- CONSTRAINT properties_type_check
    --     CHECK (type IN ('homestay', 'villa', 'guesthost'))
    CONSTRAINT properties_booking_type
         CHECK (booking_type IN ('room', 'unit'))
);

CREATE TABLE rentables (
    id VARCHAR(36) PRIMARY KEY,
    property_id VARCHAR(36) NOT NULL,
    type VARCHAR(20), -- 🔥 kunci utama : (room, unit)
    name VARCHAR(255), -- Deluxe Room / Villa A
    image_url VARCHAR(255),
    capacity INT,
    base_price DOUBLE PRECISION NOT NULL,
    discount FLOAT,
    stock INT NOT NULL DEFAULT 1, -- 🔥 pembeda utama
    created_at BIGINT,
    updated_at BIGINT,

    CONSTRAINT rentables_type
         CHECK (type IN ('room', 'unit'))
)