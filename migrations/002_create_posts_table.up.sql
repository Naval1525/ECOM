CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- If an old users table exists with integer id, convert it to UUID to match FKs
DO $$
DECLARE
    col_type text;
BEGIN
    SELECT data_type INTO col_type
    FROM information_schema.columns
    WHERE table_name = 'users' AND column_name = 'id';

    IF col_type = 'integer' THEN
        -- Drop default if any, then convert and set UUID default
        BEGIN
            ALTER TABLE users ALTER COLUMN id DROP DEFAULT;
        EXCEPTION WHEN others THEN
            -- ignore if no default
            NULL;
        END;
        ALTER TABLE users ALTER COLUMN id TYPE uuid USING uuid_generate_v4();
        ALTER TABLE users ALTER COLUMN id SET DEFAULT uuid_generate_v4();
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    image_url VARCHAR(255) DEFAULT '',
    like_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC);