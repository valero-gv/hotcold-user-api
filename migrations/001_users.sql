-- users table: primary key user_id (VARCHAR(20)) + two text fields
CREATE TABLE IF NOT EXISTS users (
    user_id        VARCHAR(20) PRIMARY KEY,
    deeplink       TEXT NOT NULL DEFAULT '',
    promo_message  TEXT NOT NULL DEFAULT ''
);

-- Primary key creates a btree index on user_id
-- Suitable for datasets with ~10M rows


