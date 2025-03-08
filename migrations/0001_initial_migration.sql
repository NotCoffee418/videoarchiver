-- +up
CREATE TABLE IF NOT EXISTS "playlists" (
    "id" INTEGER NOT NULL UNIQUE,
    "name" VARCHAR NOT NULL,
    "url" VARCHAR NOT NULL,
    "output_format" VARCHAR NOT NULL,
    "save_directory" VARCHAR NOT NULL,
    "thumbnail_base64" TEXT DEFAULT NULL,
    "is_enabled" BOOLEAN NOT NULL DEFAULT TRUE,
    "added_at" BIGINT NOT NULL DEFAULT (strftime('%s', 'now')),
    PRIMARY KEY("id")
);

CREATE TABLE IF NOT EXISTS "downloads" (
    "id" INTEGER NOT NULL UNIQUE,
    "playlist_id" INTEGER NOT NULL,
    "video_id" VARCHAR NOT NULL,
    "status" INTEGER NOT NULL,
    "format_downloaded" VARCHAR NOT NULL,
    "md5" VARCHAR,
    "last_attempt" BIGINT NOT NULL DEFAULT (strftime('%s', 'now')),
    "fail_message" VARCHAR,
    "attempt_count" INTEGER NOT NULL,
    PRIMARY KEY("id"),
    FOREIGN KEY ("playlist_id") REFERENCES "playlists"("id")
    ON UPDATE RESTRICT ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS "downloads_index_0"
ON "downloads" ("playlist_id", "video_id", "md5");

CREATE TABLE IF NOT EXISTS "logs" (
    "id" INTEGER NOT NULL UNIQUE,
    "verbosity" INTEGER NOT NULL,
    "timestamp" BIGINT NOT NULL DEFAULT (strftime('%s', 'now')),
    "message" TEXT NOT NULL,
    PRIMARY KEY("id")
);

-- +down
DROP TABLE IF EXISTS "playlists";
DROP TABLE IF EXISTS "downloads";
DROP TABLE IF EXISTS "logs";
