CREATE TYPE reservation_status AS ENUM ('active', 'cancelled');

CREATE TABLE reservations (
    id          UUID               PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID               NOT NULL REFERENCES users(id)     ON DELETE CASCADE,
    showtime_id UUID               NOT NULL REFERENCES showtimes(id) ON DELETE CASCADE,
    status      reservation_status NOT NULL DEFAULT 'active',
    total_price NUMERIC(10,2)      NOT NULL CHECK (total_price >= 0),
    created_at  TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

-- One row per reserved seat per showtime — the UNIQUE constraint is the
-- database-level guarantee against overbooking.
CREATE TABLE reservation_seats (
    reservation_id UUID NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    seat_id        UUID NOT NULL REFERENCES seats(id)        ON DELETE CASCADE,
    showtime_id    UUID NOT NULL REFERENCES showtimes(id)    ON DELETE CASCADE,
    PRIMARY KEY (reservation_id, seat_id),
    UNIQUE (seat_id, showtime_id)   -- a seat can only be reserved once per showtime
);

CREATE INDEX idx_reservations_user     ON reservations (user_id);
CREATE INDEX idx_reservations_showtime ON reservations (showtime_id);
