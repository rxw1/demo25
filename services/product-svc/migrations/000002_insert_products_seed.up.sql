INSERT INTO
    products (name, price)
VALUES
    ('Food', 10),
    ('Wood', 30),
    ('Stone', 50),
    ('Ore', 70),
    ('Coal', 90) ON CONFLICT (name) DO NOTHING;