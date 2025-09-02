-- +up
-- Indexes for playlists table
CREATE INDEX "playlists_is_enabled_index" ON "playlists" ("is_enabled");
CREATE INDEX "playlists_added_at_index" ON "playlists" ("added_at");
CREATE INDEX "playlists_duplicate_check_index" ON "playlists" ("url", "save_directory", "output_format", "is_enabled");

-- Indexes for downloads table
CREATE INDEX "downloads_status_index" ON "downloads" ("status");
CREATE INDEX "downloads_last_attempt_index" ON "downloads" ("last_attempt");
CREATE INDEX "downloads_status_last_attempt_index" ON "downloads" ("status", "last_attempt");

-- +down
DROP INDEX IF EXISTS "downloads_status_last_attempt_index";
DROP INDEX IF EXISTS "downloads_last_attempt_index";
DROP INDEX IF EXISTS "downloads_status_index";
DROP INDEX IF EXISTS "playlists_duplicate_check_index";
DROP INDEX IF EXISTS "playlists_added_at_index";
DROP INDEX IF EXISTS "playlists_is_enabled_index";