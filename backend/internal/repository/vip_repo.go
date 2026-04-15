package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type vipRepository struct {
	sql sqlExecutor
}

func NewVIPRepository(sqlDB *sql.DB) service.VIPStatsRepository {
	return &vipRepository{sql: sqlDB}
}

func (r *vipRepository) EnsureUser(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return nil
	}
	_, err := r.sql.ExecContext(ctx, `
		INSERT INTO user_vip_stats (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
	`, userID)
	return err
}

func (r *vipRepository) AddRecharge(ctx context.Context, userID int64, amount float64) error {
	if err := r.EnsureUser(ctx, userID); err != nil {
		return err
	}
	_, err := r.sql.ExecContext(ctx, `
		UPDATE user_vip_stats
		SET total_recharge_amount = total_recharge_amount + $2,
		    updated_at = NOW()
		WHERE user_id = $1
	`, userID, amount)
	return err
}

func (r *vipRepository) AddPendingSpend(ctx context.Context, userID int64, amount float64) error {
	if userID <= 0 || amount <= 0 {
		return nil
	}
	if err := r.EnsureUser(ctx, userID); err != nil {
		return err
	}
	_, err := r.sql.ExecContext(ctx, `
		INSERT INTO vip_spend_aggregation_state (user_id, pending_spend_amount)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE
		SET pending_spend_amount = vip_spend_aggregation_state.pending_spend_amount + EXCLUDED.pending_spend_amount,
		    updated_at = NOW()
	`, userID, amount)
	return err
}

func (r *vipRepository) GetSummary(ctx context.Context, userID int64) (*service.UserVIPStats, error) {
	if userID <= 0 {
		return &service.UserVIPStats{UserID: userID}, nil
	}
	stats := &service.UserVIPStats{UserID: userID}
	if err := r.EnsureUser(ctx, userID); err != nil {
		return nil, err
	}
	err := scanSingleRow(ctx, r.sql, `
		SELECT
			COALESCE(v.total_recharge_amount, 0),
			COALESCE(v.total_spend_amount, 0),
			COALESCE(s.pending_spend_amount, 0)
		FROM user_vip_stats v
		LEFT JOIN vip_spend_aggregation_state s ON s.user_id = v.user_id
		WHERE v.user_id = $1
	`, []any{userID}, &stats.RechargeTotal, &stats.SpendTotal, &stats.PendingSpend)
	if err != nil {
		return nil, fmt.Errorf("get vip summary: %w", err)
	}
	return stats, nil
}
