-- Add fallback group for API keys: auto-switch when primary group has no available accounts
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS fallback_group_id BIGINT NULL;
