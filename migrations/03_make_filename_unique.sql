BEGIN;
DELETE FROM file_metadata a
    USING file_metadata b
WHERE a.id > b.id
  AND a.filename = b.filename;
COMMIT;
DO $$
    BEGIN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_constraint
            WHERE conrelid = 'file_metadata'::regclass
              AND conname  = 'file_metadata_filename_unique'
        ) THEN
            ALTER TABLE file_metadata
                ADD CONSTRAINT file_metadata_filename_unique
                    UNIQUE (filename);
        END IF;
    END
$$;
