-- +up
CREATE TABLE IF NOT EXISTS "file_registry" (
    "id" INTEGER NOT NULL UNIQUE,
    "filename" VARCHAR NOT NULL,
    "file_path" VARCHAR NOT NULL UNIQUE,
    "md5_hash" VARCHAR NOT NULL,
    "registered_at" BIGINT NOT NULL DEFAULT (strftime('%s', 'now')),
    PRIMARY KEY("id")
);

-- Create indexes for efficient lookups
CREATE INDEX "file_registry_path_index" ON "file_registry" ("file_path");
CREATE INDEX "file_registry_md5_index" ON "file_registry" ("md5_hash");

-- +down
DROP INDEX IF EXISTS "file_registry_md5_index";
DROP INDEX IF EXISTS "file_registry_path_index";
DROP TABLE IF EXISTS "file_registry";