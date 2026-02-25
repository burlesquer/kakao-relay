-- Remove plaintext token columns (only hashes should be stored)
ALTER TABLE accounts DROP COLUMN IF EXISTS relay_token;
ALTER TABLE sessions DROP COLUMN IF EXISTS session_token;
