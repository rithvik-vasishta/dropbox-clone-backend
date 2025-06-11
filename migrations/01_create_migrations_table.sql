CREATE TABLE IF NOT EXISTS file_metadata (
     id SERIAL PRIMARY KEY,
     filename TEXT NOT NULL UNIQUE,
     num_shards INTEGER NOT NULL,
     shard_paths TEXT[] NOT NULL,
     uploaded_at TIMESTAMP DEFAULT NOW()
);