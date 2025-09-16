-- Revert creation of ULID helper and pgcrypto extension
DROP FUNCTION IF EXISTS generate_ulid();
DROP EXTENSION IF EXISTS pgcrypto;