package service

import (
	"context"
	"fmt"
	"strings"
)

type TrialCampaignBinding struct {
	RedeemCodeID   int64
	CampaignID     int64
	CampaignKey    string
	CampaignName   string
	AllowRepeat    bool
}

type TrialCampaignRepository interface {
	UpsertCampaign(ctx context.Context, campaignKey, campaignName string) (int64, error)
	BindRedeemCode(ctx context.Context, redeemCodeID, campaignID int64) error
	GetBindingByRedeemCodeID(ctx context.Context, redeemCodeID int64) (*TrialCampaignBinding, error)
	Claim(ctx context.Context, userID, redeemCodeID int64) error
}

type TrialCampaignService struct {
	repo TrialCampaignRepository
}

func NewTrialCampaignService(repo TrialCampaignRepository) *TrialCampaignService {
	return &TrialCampaignService{repo: repo}
}

func (s *TrialCampaignService) BindRedeemCode(ctx context.Context, redeemCodeID int64, campaignKey, campaignName string) error {
	if s == nil || s.repo == nil {
		return nil
	}
	campaignKey = strings.TrimSpace(campaignKey)
	if campaignKey == "" || redeemCodeID <= 0 {
		return nil
	}
	campaignID, err := s.repo.UpsertCampaign(ctx, campaignKey, strings.TrimSpace(campaignName))
	if err != nil {
		return err
	}
	return s.repo.BindRedeemCode(ctx, redeemCodeID, campaignID)
}

func (s *TrialCampaignService) DecorateRedeemCode(ctx context.Context, code *RedeemCode) error {
	if s == nil || s.repo == nil || code == nil || code.ID <= 0 {
		return nil
	}
	binding, err := s.repo.GetBindingByRedeemCodeID(ctx, code.ID)
	if err != nil || binding == nil {
		return err
	}
	code.TrialCampaignKey = &binding.CampaignKey
	if strings.TrimSpace(binding.CampaignName) != "" {
		code.TrialCampaignName = &binding.CampaignName
	}
	return nil
}

func (s *TrialCampaignService) Claim(ctx context.Context, userID int64, code *RedeemCode) error {
	if s == nil || s.repo == nil || code == nil || code.ID <= 0 {
		return nil
	}
	binding, err := s.repo.GetBindingByRedeemCodeID(ctx, code.ID)
	if err != nil || binding == nil {
		return err
	}
	if binding.AllowRepeat {
		return nil
	}
	if err := s.repo.Claim(ctx, userID, code.ID); err != nil {
		if err == ErrRedeemCodeUsed {
			return ErrTrialCampaignAlreadyClaimed
		}
		return fmt.Errorf("claim trial campaign: %w", err)
	}
	return nil
}
