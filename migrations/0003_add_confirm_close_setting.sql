-- +up
INSERT INTO "settings" (setting_key, setting_value) VALUES 
('confirm_close_enabled', 'true');

-- +down
DELETE FROM "settings" WHERE setting_key = 'confirm_close_enabled';