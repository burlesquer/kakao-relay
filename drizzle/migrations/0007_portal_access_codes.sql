-- Portal access codes: temporary login codes for portal access

CREATE TABLE "portal_access_codes" (
	"code" text PRIMARY KEY NOT NULL,
	"conversation_key" text NOT NULL,
	"expires_at" timestamp with time zone NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"used_at" timestamp with time zone,
	"last_accessed_at" timestamp with time zone
);

-- Create indexes for portal_access_codes
CREATE INDEX "portal_access_codes_expires_at_idx" ON "portal_access_codes" USING btree ("expires_at");
CREATE INDEX "portal_access_codes_conversation_key_idx" ON "portal_access_codes" USING btree ("conversation_key");
-- Composite index for finding active codes by conversation_key
CREATE INDEX "portal_access_codes_active_by_conv_idx" ON "portal_access_codes" USING btree ("conversation_key", "expires_at") WHERE "used_at" IS NULL;
