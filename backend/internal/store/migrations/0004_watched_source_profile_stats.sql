-- Real Instagram profile stats (follower/following/post counts, display
-- name, avatar, verified badge) pulled from Apify's profile-scraper actor —
-- the same actor already called for the Instagram connector's health check.
-- All nullable: a source with no successful profile check yet simply has no
-- numbers, never a fabricated placeholder.
ALTER TABLE watched_sources ADD COLUMN IF NOT EXISTS follower_count INTEGER;
ALTER TABLE watched_sources ADD COLUMN IF NOT EXISTS following_count INTEGER;
ALTER TABLE watched_sources ADD COLUMN IF NOT EXISTS post_count INTEGER;
ALTER TABLE watched_sources ADD COLUMN IF NOT EXISTS full_name TEXT;
ALTER TABLE watched_sources ADD COLUMN IF NOT EXISTS profile_pic_url TEXT;
ALTER TABLE watched_sources ADD COLUMN IF NOT EXISTS verified BOOLEAN;
ALTER TABLE watched_sources ADD COLUMN IF NOT EXISTS profile_checked_at TIMESTAMPTZ;
