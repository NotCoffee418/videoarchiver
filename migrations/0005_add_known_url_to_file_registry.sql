-- +up
ALTER TABLE file_registry ADD COLUMN known_url VARCHAR;

-- Create index for efficient YouTube URL lookups
CREATE INDEX "file_registry_known_url_index" ON "file_registry" ("known_url");

-- +down
-- Remove index
DROP INDEX IF EXISTS "file_registry_known_url_index";

-- SQLite doesn't support DROP COLUMN, so we'll need to recreate the table
-- Backup data
CREATE TEMPORARY TABLE file_registry_backup AS SELECT id, filename, file_path, md5, registered_at FROM file_registry;

-- Drop and recreate table without known_url column
DROP TABLE file_registry;
CREATE TABLE IF NOT EXISTS "file_registry" (
    "id" INTEGER NOT NULL,
    "filename" VARCHAR NOT NULL,
    "file_path" VARCHAR NOT NULL,
    "md5" VARCHAR NOT NULL,
    "registered_at" BIGINT NOT NULL DEFAULT (strftime('%s', 'now')),
    PRIMARY KEY("id"),
    UNIQUE("file_path", "md5")
);

-- Restore data
INSERT INTO file_registry (id, filename, file_path, md5, registered_at) 
SELECT id, filename, file_path, md5, registered_at FROM file_registry_backup;
DROP TABLE file_registry_backup;

-- Recreate the original index
CREATE INDEX "file_registry_md5_index" ON "file_registry" ("md5");