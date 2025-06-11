ALTER TABLE file_metadata
    RENAME COLUMN shard_paths TO primary_shards;

ALTER TABLE file_metadata
    ADD COLUMN redundant_shards TEXT[][] NOT NULL DEFAULT ARRAY[]::TEXT[][];