-- +up
CREATE TABLE IF NOT EXISTS "file_registry" (
    "id" INTEGER NOT NULL,
    "filename" VARCHAR NOT NULL,
    "file_path" VARCHAR NOT NULL,
    "md5" VARCHAR NOT NULL,
    "registered_at" BIGINT NOT NULL DEFAULT (strftime('%s', 'now')),
    PRIMARY KEY("id"),
    UNIQUE("file_path", "md5")
);

-- Create indexes for efficient lookups
CREATE INDEX "file_registry_md5_index" ON "file_registry" ("md5");

-- Add indexes to existing tables for improved query performance
-- Index for filtering active playlists
CREATE INDEX "playlists_is_enabled_index" ON "playlists" ("is_enabled");

-- Index for filtering downloads by status
CREATE INDEX "downloads_status_index" ON "downloads" ("status");

-- +down
DROP INDEX IF EXISTS "downloads_status_index";
DROP INDEX IF EXISTS "playlists_is_enabled_index";
DROP INDEX IF EXISTS "file_registry_md5_index";
DROP TABLE IF EXISTS "file_registry";