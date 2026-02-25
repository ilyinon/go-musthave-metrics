DROP TRIGGER IF EXISTS trg_gauges_updated ON gauges;
DROP TRIGGER IF EXISTS trg_counters_updated ON counters;

DROP TABLE IF EXISTS gauges;
DROP TABLE IF EXISTS counters;

DROP FUNCTION IF EXISTS set_updated_at();