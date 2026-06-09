
-- +up
ALTER TABLE downloads ADD COLUMN thumbnail_base64 TEXT DEFAULT NULL;

-- +down
ALTER TABLE downloads DROP COLUMN thumbnail_base64;
