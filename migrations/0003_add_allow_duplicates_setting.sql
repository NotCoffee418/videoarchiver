-- +up
INSERT INTO "settings" (setting_key, setting_value) VALUES 
('allow_duplicates', 'false');

-- +down
DELETE FROM "settings" WHERE setting_key = 'allow_duplicates';