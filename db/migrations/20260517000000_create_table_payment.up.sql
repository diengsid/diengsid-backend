CREATE TABLE payments (
    id          VARCHAR(36) PRIMARY KEY,
    booking_id  VARCHAR(36) NOT NULL,
    user_id     VARCHAR(36) NOT NULL,
    invoice_no  VARCHAR(100) NOT NULL UNIQUE,
    amount      BIGINT NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    payment_url TEXT,
    created_at  BIGINT NOT NULL,
    updated_at  BIGINT NOT NULL,

    CONSTRAINT fk_payments_booking
        FOREIGN KEY (booking_id)
        REFERENCES bookings(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_payments_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT payments_status_check
        CHECK (status IN ('PENDING', 'SUCCESS', 'FAILED', 'EXPIRED'))
);
