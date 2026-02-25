-- Pairing system: shared Kakao channel architecture

CREATE TYPE "public"."pairing_state" AS ENUM('unpaired', 'pending', 'paired', 'blocked');

-- Add 'acked' to inbound_message_status enum
ALTER TYPE "public"."inbound_message_status" ADD VALUE 'acked';

-- Add disabledAt to accounts
ALTER TABLE "accounts" ADD COLUMN "disabled_at" timestamp with time zone;

-- Create conversation_mappings table
CREATE TABLE "conversation_mappings" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"conversation_key" text NOT NULL UNIQUE,
	"kakao_channel_id" text NOT NULL,
	"plusfriend_user_key" text NOT NULL,
	"account_id" uuid REFERENCES "accounts"("id") ON DELETE SET NULL,
	"state" "pairing_state" DEFAULT 'unpaired' NOT NULL,
	"last_callback_url" text,
	"last_callback_expires_at" timestamp with time zone,
	"first_seen_at" timestamp with time zone DEFAULT now() NOT NULL,
	"last_seen_at" timestamp with time zone DEFAULT now() NOT NULL,
	"paired_at" timestamp with time zone
);

-- Create indexes for conversation_mappings
CREATE INDEX "conversation_mappings_account_id_idx" ON "conversation_mappings" USING btree ("account_id");
CREATE INDEX "conversation_mappings_state_idx" ON "conversation_mappings" USING btree ("state");
CREATE UNIQUE INDEX "conversation_mappings_channel_user_idx" ON "conversation_mappings" USING btree ("kakao_channel_id", "plusfriend_user_key");

-- Create pairing_codes table
CREATE TABLE "pairing_codes" (
	"code" text PRIMARY KEY NOT NULL,
	"account_id" uuid NOT NULL REFERENCES "accounts"("id") ON DELETE CASCADE,
	"expires_at" timestamp with time zone NOT NULL,
	"used_at" timestamp with time zone,
	"used_by" text,
	"metadata" jsonb,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL
);

-- Create indexes for pairing_codes
CREATE INDEX "pairing_codes_account_id_idx" ON "pairing_codes" USING btree ("account_id");
CREATE INDEX "pairing_codes_expires_at_idx" ON "pairing_codes" USING btree ("expires_at");

-- Add conversationKey to inbound_messages
ALTER TABLE "inbound_messages" ADD COLUMN "conversation_key" text;
ALTER TABLE "inbound_messages" ADD COLUMN "source_event_id" text UNIQUE;
ALTER TABLE "inbound_messages" ADD COLUMN "acked_at" timestamp with time zone;

-- Backfill conversationKey for existing messages (use a placeholder pattern)
UPDATE "inbound_messages" SET "conversation_key" = 'legacy:' || "id"::text WHERE "conversation_key" IS NULL;

-- Now make it NOT NULL
ALTER TABLE "inbound_messages" ALTER COLUMN "conversation_key" SET NOT NULL;

-- Create index for conversationKey
CREATE INDEX "inbound_messages_conversation_key_idx" ON "inbound_messages" USING btree ("conversation_key");

-- Add conversationKey to outbound_messages
ALTER TABLE "outbound_messages" ADD COLUMN "conversation_key" text;

-- Backfill conversationKey for existing outbound messages
UPDATE "outbound_messages" SET "conversation_key" = 'legacy:' || "id"::text WHERE "conversation_key" IS NULL;

-- Now make it NOT NULL
ALTER TABLE "outbound_messages" ALTER COLUMN "conversation_key" SET NOT NULL;

-- Create index for conversationKey
CREATE INDEX "outbound_messages_conversation_key_idx" ON "outbound_messages" USING btree ("conversation_key");

-- Drop old mappings table (after migrating data if needed)
-- First, migrate any existing mappings to conversation_mappings
INSERT INTO "conversation_mappings" ("conversation_key", "kakao_channel_id", "plusfriend_user_key", "account_id", "state", "last_seen_at")
SELECT 
  'default:' || "kakao_user_key",
  'default',
  "kakao_user_key",
  "account_id",
  'paired',
  "last_seen_at"
FROM "mappings"
ON CONFLICT ("conversation_key") DO NOTHING;

-- Update pairedAt for migrated mappings
UPDATE "conversation_mappings" SET "paired_at" = "first_seen_at" WHERE "state" = 'paired' AND "paired_at" IS NULL;

-- Now drop the old table
DROP TABLE "mappings";
