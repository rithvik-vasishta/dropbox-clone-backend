DO $$
    BEGIN
        IF EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_name = 'file_metadata'
              AND column_name = 'shard_paths'
        ) THEN
            ALTER TABLE file_metadata RENAME COLUMN shard_paths TO primary_shards;
        END IF;
    END$$;

DO $$
    BEGIN
        IF NOT EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_name = 'file_metadata'
              AND column_name = 'redundant_shards'
        ) THEN
            ALTER TABLE file_metadata
                ADD COLUMN redundant_shards TEXT[][] NOT NULL DEFAULT ARRAY[]::TEXT[][];
        END IF;
    END$$;
