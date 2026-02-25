CREATE TABLE IF NOT EXISTS "portal_sessions" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
  "token_hash" text NOT NULL UNIQUE,
  "user_id" uuid NOT NULL REFERENCES "portal_users"("id") ON DELETE CASCADE,
  "expires_at" timestamp with time zone NOT NULL,
  "created_at" timestamp with time zone DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS "portal_sessions_token_hash_idx" ON "portal_sessions" USING btree ("token_hash");
CREATE INDEX IF NOT EXISTS "portal_sessions_user_id_idx" ON "portal_sessions" USING btree ("user_id");
CREATE INDEX IF NOT EXISTS "portal_sessions_expires_at_idx" ON "portal_sessions" USING btree ("expires_at");
