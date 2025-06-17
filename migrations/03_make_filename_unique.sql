BEGIN;
ALTER TABLE file_metadata
    ADD CONSTRAINT file_metadata_filename_unique
        UNIQUE (filename);
COMMIT;