package service

import (
	"context"
	"strings"
)

type RedeemCodeMetadata struct {
	RedeemCodeID      int64
	CashPriceCNY      float64
	TrialCampaignKey  string
	TrialCampaignName string
}

type RedeemCodeMetadataRepository interface {
	UpsertCashPrice(ctx context.Context, redeemCodeID int64, cashPriceCNY float64) error
	GetCashPrice(ctx context.Context, redeemCodeID int64) (float64, error)
}

type RedeemMetadataService struct {
	repo RedeemCodeMetadataRepository
}

func NewRedeemMetadataService(repo RedeemCodeMetadataRepository) *RedeemMetadataService {
	return &RedeemMetadataService{repo: repo}
}

func (s *RedeemMetadataService) SetCashPrice(ctx context.Context, redeemCodeID int64, cashPriceCNY float64) error {
	if s == nil || s.repo == nil || redeemCodeID <= 0 || cashPriceCNY <= 0 {
		return nil
	}
	return s.repo.UpsertCashPrice(ctx, redeemCodeID, cashPriceCNY)
}

func (s *RedeemMetadataService) DecorateRedeemCode(ctx context.Context, code *RedeemCode) error {
	if s == nil || s.repo == nil || code == nil || code.ID <= 0 {
		return nil
	}
	cashPrice, err := s.repo.GetCashPrice(ctx, code.ID)
	if err != nil {
		return err
	}
	if cashPrice > 0 {
		code.CashPriceCNY = &cashPrice
	}
	if code.TrialCampaignKey != nil {
		value := strings.TrimSpace(*code.TrialCampaignKey)
		if value == "" {
			code.TrialCampaignKey = nil
		}
	}
	if code.TrialCampaignName != nil {
		value := strings.TrimSpace(*code.TrialCampaignName)
		if value == "" {
			code.TrialCampaignName = nil
		}
	}
	return nil
}

func (s *RedeemMetadataService) GetCashPrice(ctx context.Context, redeemCodeID int64) (float64, error) {
	if s == nil || s.repo == nil || redeemCodeID <= 0 {
		return 0, nil
	}
	return s.repo.GetCashPrice(ctx, redeemCodeID)
}
