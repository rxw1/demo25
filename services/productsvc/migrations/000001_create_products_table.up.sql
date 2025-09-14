CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS products (
    id TEXT PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4 (),
    name TEXT NOT NULL UNIQUE,
    price INT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_products_name ON products (name);
