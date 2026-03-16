CREATE TABLE showtimes (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    movie_id   UUID        NOT NULL REFERENCES movies(id)   ON DELETE CASCADE,
    theater_id UUID        NOT NULL REFERENCES theaters(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ NOT NULL,
    price      NUMERIC(10,2) NOT NULL CHECK (price >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_showtime_order CHECK (end_time > start_time)
);

-- Prevents double-booking the same theater at overlapping times.
CREATE INDEX idx_showtimes_theater_time ON showtimes (theater_id, start_time, end_time);
CREATE INDEX idx_showtimes_start_time   ON showtimes (start_time);
