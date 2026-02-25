-- Migration: Simplify for personal use
-- Drop admin/portal/oauth tables no longer needed

DROP TABLE IF EXISTS admin_sessions CASCADE;
DROP TABLE IF EXISTS portal_sessions CASCADE;
DROP TABLE IF EXISTS oauth_accounts CASCADE;
DROP TABLE IF EXISTS oauth_states CASCADE;
DROP TABLE IF EXISTS portal_users CASCADE;
DROP TABLE IF EXISTS portal_access_codes CASCADE;
DROP TABLE IF EXISTS pairing_codes CASCADE;

-- Simplify accounts table
ALTER TABLE accounts DROP COLUMN IF EXISTS mode;
ALTER TABLE accounts DROP COLUMN IF EXISTS openclaw_user_id;
ALTER TABLE accounts DROP COLUMN IF EXISTS disabled_at;
