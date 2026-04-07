-- 081_api_key_partial_unique_index.sql
-- 修复软删除 API Key 后无法再创建相同 Key 的唯一键冲突
-- 将 api_keys.key 的普通唯一约束替换为部分唯一索引（WHERE deleted_at IS NULL）
-- 软删除的记录不再占用唯一约束位置，允许重新创建相同 key

-- 删除旧的唯一约束（可能的命名方式）
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS api_keys_key_key;
DROP INDEX IF EXISTS api_keys_key_key;
DROP INDEX IF EXISTS api_key_key_key;
DROP INDEX IF EXISTS apikey_key;

-- 创建部分唯一索引：只对未删除的记录建立唯一约束
CREATE UNIQUE INDEX IF NOT EXISTS api_keys_key_unique_active
    ON api_keys(key)
    WHERE deleted_at IS NULL;
