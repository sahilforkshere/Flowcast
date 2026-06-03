-- Run this once after `docker compose up` to create the table
-- Command: docker exec -i clickhouse clickhouse-client --query "$(cat schema.sql)"

CREATE DATABASE IF NOT EXISTS analytics;

CREATE TABLE IF NOT EXISTS analytics.pageviews
(
    event_time   DateTime,
    session_id   String,
    user_id      String,
    page_url     String,
    referrer     String,
    country      String,
    device_type  String,
    duration_ms  UInt32
)
ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(event_time)
ORDER BY (event_time, session_id);
