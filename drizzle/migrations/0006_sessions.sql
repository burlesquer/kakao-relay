-- Sessions: auto-created connections from OpenClaw plugin
-- Session is initially pending_pairing, becomes paired when user enters pairing code

CREATE TYPE "public"."session_status" AS ENUM('pending_pairing', 'paired', 'expired', 'disconnected');

CREATE TABLE "sessions" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"session_token" text NOT NULL UNIQUE,
	"session_token_hash" text NOT NULL UNIQUE,
	"pairing_code" text NOT NULL UNIQUE,
	"status" "session_status" DEFAULT 'pending_pairing' NOT NULL,
	"account_id" uuid REFERENCES "accounts"("id") ON DELETE SET NULL,
	"paired_conversation_key" text,
	"metadata" jsonb,
	"expires_at" timestamp with time zone NOT NULL,
	"paired_at" timestamp with time zone,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);

-- Indexes
CREATE INDEX "sessions_session_token_hash_idx" ON "sessions" USING btree ("session_token_hash");
CREATE INDEX "sessions_pairing_code_idx" ON "sessions" USING btree ("pairing_code");
CREATE INDEX "sessions_status_idx" ON "sessions" USING btree ("status");
CREATE INDEX "sessions_expires_at_idx" ON "sessions" USING btree ("expires_at");
CREATE INDEX "sessions_account_id_idx" ON "sessions" USING btree ("account_id");
