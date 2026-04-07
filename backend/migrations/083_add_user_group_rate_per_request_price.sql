ALTER TABLE user_group_rate_multipliers
  ADD COLUMN IF NOT EXISTS per_request_price DECIMAL(20,8) NULL;
