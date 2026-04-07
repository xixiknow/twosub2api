-- 086: Restore media_type column on usage_logs
-- This column is used by image generation billing (not Sora-only),
-- and was incorrectly dropped in migration 084.
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS media_type VARCHAR(16);
