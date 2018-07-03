
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS "messages" (
    "id" integer primary key autoincrement,
    "domain_hash" varchar(255) NOT NULL,
    "message_id"  varchar(255) NOT NULL,
    "created_at"  datetime,
    "updated_at"  datetime,
    "deleted_at"  datetime,
    "spf"   varchar(255),
    "dkim"  varchar(255),
    "dmarc" varchar(255),
    "mail_server"   varchar(255),
    "error_message" varchar(1024),
    "successful" boolean,
    "mx" varchar(255));

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE "messages";
