ALTER TABLE posts
    ADD COLUMN IF NOT EXISTS is_repost BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS is_reply BOOLEAN NOT NULL DEFAULT false;

-- 既存のデータを更新
UPDATE posts SET is_repost = (repost_id IS NOT NULL);
UPDATE posts SET is_reply = (reply_to_id IS NOT NULL);