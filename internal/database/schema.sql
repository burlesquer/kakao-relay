-- Consolidated idempotent schema for kakao-relay
-- Generated from drizzle migrations 0000-0009 (final state)
-- Safe to run on both fresh and existing databases.

-- Enums (PostgreSQL has no CREATE TYPE IF NOT EXISTS, so we use DO blocks)
DO $$ BEGIN
    CREATE TYPE "public"."inbound_message_status" AS ENUM('queued', 'delivered', 'expired', 'acked');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE "public"."outbound_message_status" AS ENUM('pending', 'sent', 'failed');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE "public"."pairing_state" AS ENUM('unpaired', 'pending', 'paired', 'blocked');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE "public"."session_status" AS ENUM('pending_pairing', 'paired', 'expired', 'disconnected');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- Tables

CREATE TABLE IF NOT EXISTS "accounts" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    "relay_token_hash" text,
    "rate_limit_per_minute" integer DEFAULT 60 NOT NULL,
    "created_at" timestamp with time zone DEFAULT now() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE IF NOT EXISTS "inbound_messages" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    "account_id" uuid NOT NULL,
    "conversation_key" text NOT NULL,
    "kakao_payload" jsonb NOT NULL,
    "normalized_message" jsonb,
    "callback_url" text,
    "callback_expires_at" timestamp with time zone,
    "status" "inbound_message_status" DEFAULT 'queued' NOT NULL,
    "source_event_id" text UNIQUE,
    "created_at" timestamp with time zone DEFAULT now() NOT NULL,
    "delivered_at" timestamp with time zone,
    "acked_at" timestamp with time zone,
    CONSTRAINT "inbound_messages_account_id_accounts_id_fk"
        FOREIGN KEY ("account_id") REFERENCES "accounts"("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "outbound_messages" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    "account_id" uuid NOT NULL,
    "inbound_message_id" uuid,
    "conversation_key" text NOT NULL,
    "kakao_target" jsonb NOT NULL,
    "response_payload" jsonb NOT NULL,
    "status" "outbound_message_status" DEFAULT 'pending' NOT NULL,
    "error_message" text,
    "created_at" timestamp with time zone DEFAULT now() NOT NULL,
    "sent_at" timestamp with time zone,
    CONSTRAINT "outbound_messages_account_id_accounts_id_fk"
        FOREIGN KEY ("account_id") REFERENCES "accounts"("id") ON DELETE CASCADE,
    CONSTRAINT "outbound_messages_inbound_message_id_inbound_messages_id_fk"
        FOREIGN KEY ("inbound_message_id") REFERENCES "inbound_messages"("id") ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS "conversation_mappings" (
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

CREATE TABLE IF NOT EXISTS "sessions" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
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

-- Indexes: accounts
CREATE UNIQUE INDEX IF NOT EXISTS "accounts_relay_token_hash_idx"
    ON "accounts" USING btree ("relay_token_hash");

-- Indexes: inbound_messages
CREATE INDEX IF NOT EXISTS "inbound_messages_account_id_idx"
    ON "inbound_messages" USING btree ("account_id");
CREATE INDEX IF NOT EXISTS "inbound_messages_status_idx"
    ON "inbound_messages" USING btree ("status");
CREATE INDEX IF NOT EXISTS "inbound_messages_created_at_idx"
    ON "inbound_messages" USING btree ("created_at");
CREATE INDEX IF NOT EXISTS "inbound_messages_conversation_key_idx"
    ON "inbound_messages" USING btree ("conversation_key");

-- Indexes: outbound_messages
CREATE INDEX IF NOT EXISTS "outbound_messages_account_id_idx"
    ON "outbound_messages" USING btree ("account_id");
CREATE INDEX IF NOT EXISTS "outbound_messages_inbound_message_id_idx"
    ON "outbound_messages" USING btree ("inbound_message_id");
CREATE INDEX IF NOT EXISTS "outbound_messages_status_idx"
    ON "outbound_messages" USING btree ("status");
CREATE INDEX IF NOT EXISTS "outbound_messages_conversation_key_idx"
    ON "outbound_messages" USING btree ("conversation_key");

-- Indexes: conversation_mappings
CREATE INDEX IF NOT EXISTS "conversation_mappings_account_id_idx"
    ON "conversation_mappings" USING btree ("account_id");
CREATE INDEX IF NOT EXISTS "conversation_mappings_state_idx"
    ON "conversation_mappings" USING btree ("state");
CREATE UNIQUE INDEX IF NOT EXISTS "conversation_mappings_channel_user_idx"
    ON "conversation_mappings" USING btree ("kakao_channel_id", "plusfriend_user_key");

-- Indexes: sessions
CREATE INDEX IF NOT EXISTS "sessions_session_token_hash_idx"
    ON "sessions" USING btree ("session_token_hash");
CREATE INDEX IF NOT EXISTS "sessions_pairing_code_idx"
    ON "sessions" USING btree ("pairing_code");
CREATE INDEX IF NOT EXISTS "sessions_status_idx"
    ON "sessions" USING btree ("status");
CREATE INDEX IF NOT EXISTS "sessions_expires_at_idx"
    ON "sessions" USING btree ("expires_at");
CREATE INDEX IF NOT EXISTS "sessions_account_id_idx"
    ON "sessions" USING btree ("account_id");
