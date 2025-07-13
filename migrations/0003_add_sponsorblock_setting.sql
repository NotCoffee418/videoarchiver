-- +up
-- Decent defaults
INSERT INTO settings (setting_key, setting_value) VALUES ('sponsorblock_video', 'sponsor,intro,outro,selfpromo,interaction,preview,filler');
INSERT INTO settings (setting_key, setting_value) VALUES ('sponsorblock_audio', 'sponsor,selfpromo,interaction,preview,filler');
-- +down
DELETE FROM settings WHERE setting_key = 'sponsorblock_video';
DELETE FROM settings WHERE setting_key = 'sponsorblock_audio';