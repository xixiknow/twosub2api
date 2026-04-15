package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type trialCampaignRepository struct {
	sql sqlExecutor
}

func NewTrialCampaignRepository(sqlDB *sql.DB) service.TrialCampaignRepository {
	return &trialCampaignRepository{sql: sqlDB}
}

func (r *trialCampaignRepository) UpsertCampaign(ctx context.Context, campaignKey, campaignName string) (int64, error) {
	var id int64
	err := scanSingleRow(ctx, r.sql, `
		INSERT INTO redeem_code_trial_campaigns (campaign_key, campaign_name)
		VALUES ($1, $2)
		ON CONFLICT (campaign_key) DO UPDATE
		SET campaign_name = CASE
			WHEN EXCLUDED.campaign_name <> '' THEN EXCLUDED.campaign_name
			ELSE redeem_code_trial_campaigns.campaign_name
		END,
		    updated_at = NOW()
		RETURNING id
	`, []any{campaignKey, campaignName}, &id)
	return id, err
}

func (r *trialCampaignRepository) BindRedeemCode(ctx context.Context, redeemCodeID, campaignID int64) error {
	_, err := r.sql.ExecContext(ctx, `
		INSERT INTO redeem_code_trial_campaign_bindings (redeem_code_id, campaign_id)
		VALUES ($1, $2)
		ON CONFLICT (redeem_code_id) DO UPDATE
		SET campaign_id = EXCLUDED.campaign_id
	`, redeemCodeID, campaignID)
	return err
}

func (r *trialCampaignRepository) GetBindingByRedeemCodeID(ctx context.Context, redeemCodeID int64) (*service.TrialCampaignBinding, error) {
	result := &service.TrialCampaignBinding{RedeemCodeID: redeemCodeID}
	err := scanSingleRow(ctx, r.sql, `
		SELECT
			c.id,
			c.campaign_key,
			c.campaign_name,
			c.allow_repeat_redeem
		FROM redeem_code_trial_campaign_bindings b
		JOIN redeem_code_trial_campaigns c ON c.id = b.campaign_id
		WHERE b.redeem_code_id = $1
	`, []any{redeemCodeID}, &result.CampaignID, &result.CampaignKey, &result.CampaignName, &result.AllowRepeat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

func (r *trialCampaignRepository) Claim(ctx context.Context, userID, redeemCodeID int64) error {
	_, err := r.sql.ExecContext(ctx, `
		INSERT INTO redeem_code_trial_campaign_claims (user_id, campaign_id, redeem_code_id)
		SELECT $1, campaign_id, $2
		FROM redeem_code_trial_campaign_bindings
		WHERE redeem_code_id = $2
	`, userID, redeemCodeID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return service.ErrRedeemCodeUsed
		}
		return fmt.Errorf("claim trial campaign: %w", err)
	}
	return nil
}
