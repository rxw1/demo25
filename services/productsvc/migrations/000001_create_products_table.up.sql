--CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
\i pgulid.sql
--CREATE EXTENSION ulid;

CREATE TABLE IF NOT EXISTS products (
    id TEXT PRIMARY KEY NOT NULL DEFAULT generate_ulid(),
    name TEXT NOT NULL UNIQUE,
    price INT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_products_name ON products (name);
