package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type redeemMetadataRepository struct {
	sql sqlExecutor
}

func NewRedeemMetadataRepository(sqlDB *sql.DB) service.RedeemCodeMetadataRepository {
	return &redeemMetadataRepository{sql: sqlDB}
}

func (r *redeemMetadataRepository) UpsertCashPrice(ctx context.Context, redeemCodeID int64, cashPriceCNY float64) error {
	if r == nil || r.sql == nil || redeemCodeID <= 0 {
		return nil
	}
	_, err := r.sql.ExecContext(ctx, `
		INSERT INTO redeem_code_cash_metadata (redeem_code_id, cash_price_cny)
		VALUES ($1, $2)
		ON CONFLICT (redeem_code_id) DO UPDATE
		SET cash_price_cny = EXCLUDED.cash_price_cny,
		    updated_at = NOW()
	`, redeemCodeID, cashPriceCNY)
	return err
}

func (r *redeemMetadataRepository) GetCashPrice(ctx context.Context, redeemCodeID int64) (float64, error) {
	if r == nil || r.sql == nil || redeemCodeID <= 0 {
		return 0, nil
	}
	var cashPrice float64
	err := scanSingleRow(ctx, r.sql, `
		SELECT cash_price_cny
		FROM redeem_code_cash_metadata
		WHERE redeem_code_id = $1
	`, []any{redeemCodeID}, &cashPrice)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return cashPrice, nil
}
