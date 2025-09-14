-- Revert seeded products inserted by 000002_insert_products_seed.up.sql
DELETE FROM products
WHERE
    name IN ('Food', 'Wood', 'Stone', 'Ore', 'Coal');