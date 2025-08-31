-- +up
INSERT INTO "settings" (setting_key, setting_value) VALUES 
('legal_disclaimer_accepted', 'false');

-- +down
DELETE FROM "settings" WHERE setting_key = 'legal_disclaimer_accepted';