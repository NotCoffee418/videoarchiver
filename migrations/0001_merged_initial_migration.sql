
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


CREATE TABLE "downloads" (
    "id" INTEGER NOT NULL UNIQUE,
    "playlist_id" INTEGER NOT NULL,
    "url" VARCHAR NOT NULL,
    "status" INTEGER NOT NULL,
    "format_downloaded" VARCHAR NOT NULL,
    "md5" VARCHAR,
	"output_filename" VARCHAR,
    "last_attempt" BIGINT NOT NULL DEFAULT (strftime('%s', 'now')),
    "fail_message" VARCHAR,
    "attempt_count" INTEGER NOT NULL,
    PRIMARY KEY("id"),
    FOREIGN KEY ("playlist_id") REFERENCES "playlists"("id")
    ON UPDATE RESTRICT ON DELETE RESTRICT
);

CREATE INDEX "downloads_index_0"
ON "downloads" ("playlist_id", "url", "md5");

CREATE TABLE "settings" (
    "setting_key" VARCHAR NOT NULL UNIQUE,
    "setting_value" VARCHAR NOT NULL
);

INSERT INTO "settings" (setting_key, setting_value) VALUES 
('autostart_service', 'true'),
('autoupdate_ytdlp', 'true'),
('sponsorblock_video', 'sponsor,intro,outro,selfpromo,interaction,preview,filler'),
('sponsorblock_audio', 'sponsor,selfpromo,interaction,preview,filler'),
('daemon_signal', '0');


-- +down
DROP TABLE IF EXISTS "downloads";
DROP TABLE IF EXISTS "playlists";
DROP TABLE IF EXISTS "settings";