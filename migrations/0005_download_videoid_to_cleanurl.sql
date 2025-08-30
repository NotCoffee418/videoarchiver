-- +up
ALTER TABLE downloads RENAME COLUMN video_id TO url;

-- +down
ALTER TABLE downloads RENAME COLUMN url TO videoid;