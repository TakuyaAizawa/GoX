ALTER TABLE posts
    DROP COLUMN IF EXISTS is_repost,
    DROP COLUMN IF EXISTS is_reply;