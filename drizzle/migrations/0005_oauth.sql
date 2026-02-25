-- Clean up existing users (OAuth-only going forward)
DELETE FROM "portal_sessions";
DELETE FROM "portal_users";

-- Remove password_hash column (OAuth-only authentication)
ALTER TABLE "portal_users" DROP COLUMN "password_hash";

-- OAuth accounts table for linking external identity providers
CREATE TABLE "oauth_accounts" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    "user_id" uuid NOT NULL REFERENCES "portal_users"("id") ON DELETE CASCADE,
    "provider" text NOT NULL,
    "provider_user_id" text NOT NULL,
    "email" text,
    "access_token" text,
    "refresh_token" text,
    "token_expires_at" timestamp with time zone,
    "raw_data" jsonb,
    "created_at" timestamp with time zone DEFAULT now() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT "oauth_accounts_provider_user_unique" UNIQUE ("provider", "provider_user_id")
);

CREATE INDEX "oauth_accounts_user_id_idx" ON "oauth_accounts" USING btree ("user_id");
CREATE INDEX "oauth_accounts_provider_email_idx" ON "oauth_accounts" USING btree ("provider", "email");

-- OAuth state table for CSRF protection
CREATE TABLE "oauth_states" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    "state" text NOT NULL UNIQUE,
    "provider" text NOT NULL,
    "code_verifier" text,
    "redirect_url" text,
    "expires_at" timestamp with time zone NOT NULL,
    "created_at" timestamp with time zone DEFAULT now() NOT NULL
);

CREATE INDEX "oauth_states_state_idx" ON "oauth_states" USING btree ("state");
CREATE INDEX "oauth_states_expires_at_idx" ON "oauth_states" USING btree ("expires_at");
