CREATE TABLE IF NOT EXISTS user_vip_stats (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    total_recharge_amount NUMERIC(20,8) NOT NULL DEFAULT 0,
    total_spend_amount NUMERIC(20,8) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS vip_spend_aggregation_state (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    pending_spend_amount NUMERIC(20,8) NOT NULL DEFAULT 0,
    last_consumed_usage_log_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS redeem_code_trial_campaigns (
    id BIGSERIAL PRIMARY KEY,
    campaign_key VARCHAR(128) NOT NULL UNIQUE,
    campaign_name VARCHAR(255) NOT NULL DEFAULT '',
    allow_repeat_redeem BOOLEAN NOT NULL DEFAULT FALSE,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS redeem_code_trial_campaign_bindings (
    redeem_code_id BIGINT PRIMARY KEY REFERENCES redeem_codes(id) ON DELETE CASCADE,
    campaign_id BIGINT NOT NULL REFERENCES redeem_code_trial_campaigns(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS redeem_code_cash_metadata (
    redeem_code_id BIGINT PRIMARY KEY REFERENCES redeem_codes(id) ON DELETE CASCADE,
    cash_price_cny NUMERIC(20,8) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_redeem_code_trial_campaign_bindings_campaign_id
    ON redeem_code_trial_campaign_bindings(campaign_id);

CREATE TABLE IF NOT EXISTS redeem_code_trial_campaign_claims (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    campaign_id BIGINT NOT NULL REFERENCES redeem_code_trial_campaigns(id) ON DELETE CASCADE,
    redeem_code_id BIGINT NOT NULL REFERENCES redeem_codes(id) ON DELETE CASCADE,
    claimed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, campaign_id)
);

CREATE INDEX IF NOT EXISTS idx_redeem_code_trial_campaign_claims_campaign_id
    ON redeem_code_trial_campaign_claims(campaign_id);

ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS vip_level_code VARCHAR(64),
    ADD COLUMN IF NOT EXISTS vip_level_name VARCHAR(128),
    ADD COLUMN IF NOT EXISTS vip_base_multiplier NUMERIC(20,8),
    ADD COLUMN IF NOT EXISTS vip_final_multiplier NUMERIC(20,8),
    ADD COLUMN IF NOT EXISTS vip_discount_amount NUMERIC(20,8),
    ADD COLUMN IF NOT EXISTS vip_original_cost NUMERIC(20,8),
    ADD COLUMN IF NOT EXISTS vip_rule_key VARCHAR(255);
