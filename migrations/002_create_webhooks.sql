CREATE TABLE IF NOT EXISTS webhooks (
    id                SERIAL PRIMARY KEY,
    url               TEXT           NOT NULL,
    from_currency     VARCHAR(10)    NOT NULL,
    to_currency       VARCHAR(10)    NOT NULL,
    threshold         NUMERIC(18, 8) NOT NULL,
    direction         VARCHAR(5)     NOT NULL CHECK (direction IN ('above', 'below')),
    last_triggered_at TIMESTAMPTZ,
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);
