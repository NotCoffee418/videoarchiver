-- +up
INSERT INTO "settings" (setting_key, setting_value) VALUES 
('browser_credentials_source', 'none');

-- +down
DELETE FROM "settings" WHERE setting_key = 'browser_credentials_source';
