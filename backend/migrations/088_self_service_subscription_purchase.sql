ALTER TABLE groups ADD COLUMN IF NOT EXISTS subscription_price DECIMAL(10,2) NULL;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS subscription_display_name VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE groups ADD COLUMN IF NOT EXISTS subscription_visible BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS subscription_features JSONB NOT NULL DEFAULT '[]'::jsonb;

CREATE TABLE IF NOT EXISTS subscription_orders (
    id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(64) NOT NULL UNIQUE,
    trade_no VARCHAR(128),
    amount DECIMAL(10,2) NOT NULL,
    original_price DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    payment_method VARCHAR(32) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    notify_data TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    activated_at TIMESTAMPTZ,
    subscription_id BIGINT,
    group_id BIGINT NOT NULL REFERENCES groups(id),
    user_id BIGINT NOT NULL REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_subscription_orders_user_id ON subscription_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_subscription_orders_status ON subscription_orders(status);
CREATE INDEX IF NOT EXISTS idx_subscription_orders_order_no ON subscription_orders(order_no);
