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

-- Fix schema issues in existing tables (remove redundant constraints)
-- Remove redundant UNIQUE constraints from PRIMARY KEY columns
-- Note: SQLite doesn't support DROP CONSTRAINT, so we'll recreate the settings table
-- to make setting_key a proper PRIMARY KEY instead of just UNIQUE

-- Backup settings data
CREATE TEMPORARY TABLE settings_backup AS SELECT * FROM settings;

-- Drop and recreate settings table with proper PRIMARY KEY
DROP TABLE settings;
CREATE TABLE "settings" (
    "setting_key" VARCHAR NOT NULL,
    "setting_value" VARCHAR NOT NULL,
    PRIMARY KEY("setting_key")
);

-- Restore settings data
INSERT INTO settings SELECT * FROM settings_backup;
DROP TABLE settings_backup;

-- +down
-- Revert settings table changes (restore original structure with UNIQUE constraint)
CREATE TEMPORARY TABLE settings_backup AS SELECT * FROM settings;
DROP TABLE settings;
CREATE TABLE "settings" (
    "setting_key" VARCHAR NOT NULL UNIQUE,
    "setting_value" VARCHAR NOT NULL
);
INSERT INTO settings SELECT * FROM settings_backup;
DROP TABLE settings_backup;

-- Remove indexes
DROP INDEX IF EXISTS "downloads_status_index";
DROP INDEX IF EXISTS "playlists_is_enabled_index";
DROP INDEX IF EXISTS "file_registry_md5_index";
DROP TABLE IF EXISTS "file_registry";