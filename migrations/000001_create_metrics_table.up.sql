CREATE TABLE IF NOT EXISTS gauges (
    name  TEXT PRIMARY KEY,
    value DOUBLE PRECISION NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS counters (
    name  TEXT PRIMARY KEY,
    value BIGINT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_gauges_updated
BEFORE UPDATE ON gauges
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_counters_updated
BEFORE UPDATE ON counters
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();