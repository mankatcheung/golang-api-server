CREATE TABLE IF NOT EXISTS logs (
    id         BIGSERIAL    PRIMARY KEY,
    logged_at  TIMESTAMPTZ  NOT NULL,
    level      VARCHAR(10)  NOT NULL,
    message    TEXT         NOT NULL,
    attrs      JSONB,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_logs_logged_at ON logs (logged_at DESC);
CREATE INDEX IF NOT EXISTS idx_logs_level     ON logs (level);
CREATE INDEX IF NOT EXISTS idx_logs_attrs     ON logs USING gin (attrs);
