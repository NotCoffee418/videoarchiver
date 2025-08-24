-- +up
ALTER TABLE playlists RENAME COLUMN url TO playlist_guid;
-- +down
ALTER TABLE playlists RENAME COLUMN playlist_guid TO url;