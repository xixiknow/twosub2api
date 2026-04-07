-- Migration: 084_drop_sora_tables_and_columns
-- Remove all Sora-related tables and columns introduced by migrations 046, 047, 063.

-- ============================================================
-- 1. Drop Sora tables
-- ============================================================
DROP TABLE IF EXISTS sora_generations;
DROP TABLE IF EXISTS sora_accounts;

-- ============================================================
-- 2. Drop Sora columns from groups table (added by 047, 063)
-- ============================================================
ALTER TABLE groups
    DROP COLUMN IF EXISTS sora_image_price_360,
    DROP COLUMN IF EXISTS sora_image_price_540,
    DROP COLUMN IF EXISTS sora_video_price_per_request,
    DROP COLUMN IF EXISTS sora_video_price_per_request_hd,
    DROP COLUMN IF EXISTS sora_storage_quota_bytes;

-- ============================================================
-- 3. Drop Sora columns from users table (added by 063)
-- ============================================================
ALTER TABLE users
    DROP COLUMN IF EXISTS sora_storage_quota_bytes,
    DROP COLUMN IF EXISTS sora_storage_used_bytes;

-- ============================================================
-- 4. Drop media_type column from usage_logs (added by 047)
-- ============================================================
ALTER TABLE usage_logs
    DROP COLUMN IF EXISTS media_type;
