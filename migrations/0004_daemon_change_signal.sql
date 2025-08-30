-- +up
INSERT INTO settings (setting_key, setting_value) VALUES ('daemon_signal', '0');

-- +down
DELETE FROM settings WHERE setting_key = 'daemon_signal';