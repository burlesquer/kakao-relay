-- Portal users table for self-service pairing code generation

CREATE TABLE "portal_users" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"email" text NOT NULL UNIQUE,
	"password_hash" text NOT NULL,
	"account_id" uuid NOT NULL REFERENCES "accounts"("id") ON DELETE CASCADE,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"last_login_at" timestamp with time zone
);

CREATE UNIQUE INDEX "portal_users_email_idx" ON "portal_users" USING btree ("email");
CREATE INDEX "portal_users_account_id_idx" ON "portal_users" USING btree ("account_id");
