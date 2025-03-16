-- +up
CREATE TABLE IF NOT EXISTS "settings" (
    "setting_key" VARCHAR NOT NULL UNIQUE,
    "setting_value" VARCHAR NOT NULL
);

INSERT INTO "settings" (setting_key, setting_value) VALUES 
('autostart_service', 'true'),
('autoupdate_ytdlp', 'true');

-- +down
DROP TABLE IF EXISTS "settings";