CREATE TYPE "public"."account_mode" AS ENUM('direct', 'relay');--> statement-breakpoint
CREATE TYPE "public"."inbound_message_status" AS ENUM('queued', 'delivered', 'expired');--> statement-breakpoint
CREATE TYPE "public"."outbound_message_status" AS ENUM('pending', 'sent', 'failed');--> statement-breakpoint
CREATE TABLE "accounts" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"openclaw_user_id" text,
	"relay_token" text,
	"relay_token_hash" text,
	"mode" "account_mode" DEFAULT 'relay' NOT NULL,
	"rate_limit_per_minute" integer DEFAULT 60 NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "inbound_messages" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"account_id" uuid NOT NULL,
	"kakao_payload" jsonb NOT NULL,
	"normalized_message" jsonb,
	"callback_url" text,
	"callback_expires_at" timestamp with time zone,
	"status" "inbound_message_status" DEFAULT 'queued' NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"delivered_at" timestamp with time zone
);
--> statement-breakpoint
CREATE TABLE "mappings" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"kakao_user_key" text NOT NULL,
	"account_id" uuid NOT NULL,
	"last_seen_at" timestamp with time zone DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "outbound_messages" (
	"id" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"account_id" uuid NOT NULL,
	"inbound_message_id" uuid,
	"kakao_target" jsonb NOT NULL,
	"response_payload" jsonb NOT NULL,
	"status" "outbound_message_status" DEFAULT 'pending' NOT NULL,
	"error_message" text,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"sent_at" timestamp with time zone
);
--> statement-breakpoint
ALTER TABLE "inbound_messages" ADD CONSTRAINT "inbound_messages_account_id_accounts_id_fk" FOREIGN KEY ("account_id") REFERENCES "public"."accounts"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "mappings" ADD CONSTRAINT "mappings_account_id_accounts_id_fk" FOREIGN KEY ("account_id") REFERENCES "public"."accounts"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "outbound_messages" ADD CONSTRAINT "outbound_messages_account_id_accounts_id_fk" FOREIGN KEY ("account_id") REFERENCES "public"."accounts"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "outbound_messages" ADD CONSTRAINT "outbound_messages_inbound_message_id_inbound_messages_id_fk" FOREIGN KEY ("inbound_message_id") REFERENCES "public"."inbound_messages"("id") ON DELETE set null ON UPDATE no action;--> statement-breakpoint
CREATE UNIQUE INDEX "accounts_relay_token_hash_idx" ON "accounts" USING btree ("relay_token_hash");--> statement-breakpoint
CREATE INDEX "accounts_openclaw_user_id_idx" ON "accounts" USING btree ("openclaw_user_id");--> statement-breakpoint
CREATE INDEX "inbound_messages_account_id_idx" ON "inbound_messages" USING btree ("account_id");--> statement-breakpoint
CREATE INDEX "inbound_messages_status_idx" ON "inbound_messages" USING btree ("status");--> statement-breakpoint
CREATE INDEX "inbound_messages_created_at_idx" ON "inbound_messages" USING btree ("created_at");--> statement-breakpoint
CREATE INDEX "mappings_account_id_idx" ON "mappings" USING btree ("account_id");--> statement-breakpoint
CREATE INDEX "mappings_kakao_user_key_idx" ON "mappings" USING btree ("kakao_user_key");--> statement-breakpoint
CREATE UNIQUE INDEX "mappings_account_kakao_user_idx" ON "mappings" USING btree ("account_id","kakao_user_key");--> statement-breakpoint
CREATE INDEX "outbound_messages_account_id_idx" ON "outbound_messages" USING btree ("account_id");--> statement-breakpoint
CREATE INDEX "outbound_messages_inbound_message_id_idx" ON "outbound_messages" USING btree ("inbound_message_id");--> statement-breakpoint
CREATE INDEX "outbound_messages_status_idx" ON "outbound_messages" USING btree ("status");