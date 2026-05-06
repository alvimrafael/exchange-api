CREATE TABLE IF NOT EXISTS rates (
                                     id         SERIAL PRIMARY KEY,
                                     from_currency VARCHAR(10)    NOT NULL,
    to_currency   VARCHAR(10)    NOT NULL,
    rate          NUMERIC(18, 8) NOT NULL,
    cached        BOOLEAN        NOT NULL DEFAULT FALSE,
    queried_at    TIMESTAMP      NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_rates_currencies ON rates (from_currency, to_currency);
CREATE INDEX IF NOT EXISTS idx_rates_queried_at ON rates (queried_at);