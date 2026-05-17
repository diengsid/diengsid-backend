CREATE TABLE bookings (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(36) NOT NULL,
    property_id     VARCHAR(36) NOT NULL,
    rentable_id     VARCHAR(36) NOT NULL,
    quantity        INT DEFAULT 1, -- total kamar
    check_in        DATE NOT NULL,
    check_out       DATE NOT NULL,
    total_night     INT NOT NULL,
    total_price     DOUBLE PRECISION NOT NULL,
    discount        FLOAT DEFAULT 0,
    status          VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    payment_status  VARCHAR(20) NOT NULL DEFAULT 'UNPAID',
    first_payment   VARCHAR(10),
    created_at      BIGINT NOT NULL,
    updated_at      BIGINT NOT NULL,

    CONSTRAINT fk_bookings_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_bookings_property
        FOREIGN KEY (property_id)
        REFERENCES properties(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_bookings_rentable
        FOREIGN KEY (rentable_id)
        REFERENCES rentables(id)
        ON DELETE CASCADE,

    CONSTRAINT bookings_status_check
        CHECK (status IN (
            'PENDING',
            'WAITING_PAYMENT',
            'UNAVAILABLE',
            'CANCELLED',
            'CHECK_IN',
            'REVIEW',
            'DONE'
        )),

    CONSTRAINT bookings_payment_status_check
        CHECK (payment_status IN ('UNPAID', 'PAID', 'REFUNDED')),

    CONSTRAINT bookings_first_payment_check
        CHECK (first_payment IN ('DP', 'FULL'))
);
