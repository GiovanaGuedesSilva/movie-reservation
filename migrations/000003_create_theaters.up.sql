CREATE TABLE theaters (
    id           UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name         VARCHAR(150) NOT NULL UNIQUE,
    total_rows   INTEGER      NOT NULL CHECK (total_rows > 0),
    seats_per_row INTEGER     NOT NULL CHECK (seats_per_row > 0),
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE seats (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    theater_id UUID        NOT NULL REFERENCES theaters(id) ON DELETE CASCADE,
    row        VARCHAR(5)  NOT NULL,  -- e.g. "A", "B", ..., "Z"
    number     INTEGER     NOT NULL CHECK (number > 0),
    UNIQUE (theater_id, row, number)
);

CREATE INDEX idx_seats_theater ON seats (theater_id);
