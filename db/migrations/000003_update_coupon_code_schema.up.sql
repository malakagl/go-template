-- remove old coupon codes table and files table
DROP TABLE IF EXISTS coupon_codes;
DROP TABLE IF EXISTS coupon_code_files;

-- Create new files table and coupon codes table with updated schema
CREATE TABLE files
(
    id         SERIAL PRIMARY KEY,
    file_name  VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Coupon codes table
CREATE TABLE coupon_codes
(
    id         SERIAL PRIMARY KEY,
    code       VARCHAR(10) NOT NULL,
    file_id    INT         NOT NULL REFERENCES files (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for faster coupon lookups by code
CREATE INDEX idx_coupon_codes_code ON coupon_codes (code);
