-- +up
INSERT INTO "settings" (setting_key, setting_value) VALUES 
('direct_download_last_path', ''),
('direct_download_last_format', 'mp4');

-- +down
DELETE FROM "settings" WHERE setting_key = 'direct_download_last_path';
DELETE FROM "settings" WHERE setting_key = 'direct_download_last_format';
