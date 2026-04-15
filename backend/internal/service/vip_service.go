package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

type VIPStatsRepository interface {
	EnsureUser(ctx context.Context, userID int64) error
	AddRecharge(ctx context.Context, userID int64, amount float64) error
	AddPendingSpend(ctx context.Context, userID int64, amount float64) error
	GetSummary(ctx context.Context, userID int64) (*UserVIPStats, error)
}

type UserVIPStats struct {
	UserID        int64
	RechargeTotal float64
	SpendTotal    float64
	PendingSpend  float64
}

type VIPRule struct {
	LevelCode        string             `json:"level_code"`
	LevelName        string             `json:"level_name"`
	RequiredRecharge float64            `json:"required_recharge"`
	RequiredSpend    float64            `json:"required_spend"`
	Multiplier       float64            `json:"multiplier"`
	ModelMultipliers map[string]float64 `json:"model_multipliers"`
	RuleKey          string             `json:"rule_key"`
}

type VIPPricingDecision struct {
	Applied          bool
	LevelCode        string
	LevelName        string
	BaseMultiplier   float64
	FinalMultiplier  float64
	DiscountAmount   float64
	OriginalCost     float64
	RuleKey          string
}

type VIPService struct {
	settingRepo SettingRepository
	statsRepo   VIPStatsRepository
}

func NewVIPService(settingRepo SettingRepository, statsRepo VIPStatsRepository) *VIPService {
	return &VIPService{
		settingRepo: settingRepo,
		statsRepo:   statsRepo,
	}
}

func (s *VIPService) IsEnabled(ctx context.Context) bool {
	if s == nil || s.settingRepo == nil {
		return false
	}
	value, err := s.settingRepo.GetValue(ctx, SettingKeyVIPEnabled)
	if err != nil {
		return false
	}
	return value == "true"
}

func (s *VIPService) GetRules(ctx context.Context) ([]VIPRule, error) {
	if s == nil || s.settingRepo == nil {
		return nil, nil
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyVIPRules)
	if err != nil || strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var rules []VIPRule
	if err := json.Unmarshal([]byte(raw), &rules); err != nil {
		return nil, fmt.Errorf("parse vip rules: %w", err)
	}
	for i := range rules {
		if strings.TrimSpace(rules[i].LevelCode) == "" {
			rules[i].LevelCode = fmt.Sprintf("vip_%d", i)
		}
		if strings.TrimSpace(rules[i].LevelName) == "" {
			rules[i].LevelName = strings.ToUpper(rules[i].LevelCode)
		}
		if rules[i].Multiplier <= 0 {
			rules[i].Multiplier = 1
		}
		if strings.TrimSpace(rules[i].RuleKey) == "" {
			rules[i].RuleKey = rules[i].LevelCode
		}
	}
	sort.SliceStable(rules, func(i, j int) bool {
		if rules[i].RequiredRecharge == rules[j].RequiredRecharge {
			return rules[i].RequiredSpend < rules[j].RequiredSpend
		}
		return rules[i].RequiredRecharge < rules[j].RequiredRecharge
	})
	return rules, nil
}

func (s *VIPService) ResolveUserVIP(ctx context.Context, userID int64) (*VIPSummary, *VIPNextLevel, []VIPRule, error) {
	rules, err := s.GetRules(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	enabled := s.IsEnabled(ctx) && len(rules) > 0
	stats := &UserVIPStats{UserID: userID}
	if s != nil && s.statsRepo != nil && userID > 0 {
		if err := s.statsRepo.EnsureUser(ctx, userID); err == nil {
			if loaded, loadErr := s.statsRepo.GetSummary(ctx, userID); loadErr == nil && loaded != nil {
				stats = loaded
			}
		}
	}

	current := &VIPSummary{
		Enabled:        enabled,
		LevelCode:      "vip0",
		LevelName:      "VIP0",
		BaseMultiplier: 1,
		RechargeTotal:  stats.RechargeTotal,
		SpendTotal:     stats.SpendTotal + stats.PendingSpend,
	}

	var next *VIPNextLevel
	for i := range rules {
		rule := rules[i]
		if current.RechargeTotal >= rule.RequiredRecharge && current.SpendTotal >= rule.RequiredSpend {
			current.LevelCode = rule.LevelCode
			current.LevelName = rule.LevelName
			current.BaseMultiplier = rule.Multiplier
			continue
		}
		next = &VIPNextLevel{
			LevelCode:            rule.LevelCode,
			LevelName:            rule.LevelName,
			RequiredRecharge:     rule.RequiredRecharge,
			RequiredSpend:        rule.RequiredSpend,
			RemainingRecharge:    math.Max(0, rule.RequiredRecharge-current.RechargeTotal),
			RemainingSpend:       math.Max(0, rule.RequiredSpend-current.SpendTotal),
			UnlockConditionLabel: buildVIPUnlockConditionLabel(rule),
		}
		break
	}
	if next == nil && len(rules) > 0 {
		current.ProgressPercent = 100
	} else if next != nil {
		current.ProgressPercent = computeVIPProgress(current, next)
	}

	return current, next, rules, nil
}

func (s *VIPService) ApplyCost(ctx context.Context, userID int64, model string, cost *CostBreakdown, baseMultiplier float64) (*VIPPricingDecision, error) {
	if cost == nil {
		return &VIPPricingDecision{BaseMultiplier: 1, FinalMultiplier: baseMultiplier}, nil
	}
	current, _, rules, err := s.ResolveUserVIP(ctx, userID)
	if err != nil {
		return nil, err
	}
	decision := &VIPPricingDecision{
		BaseMultiplier: 1,
		FinalMultiplier: baseMultiplier,
	}
	if current == nil || !current.Enabled {
		return decision, nil
	}
	decision.LevelCode = current.LevelCode
	decision.LevelName = current.LevelName
	decision.BaseMultiplier = current.BaseMultiplier

	selectedMultiplier := current.BaseMultiplier
	ruleKey := current.LevelCode
	for i := range rules {
		rule := rules[i]
		if rule.LevelCode != current.LevelCode {
			continue
		}
		ruleKey = rule.RuleKey
		if modelMultiplier := resolveVIPModelMultiplier(rule.ModelMultipliers, model); modelMultiplier > 0 {
			selectedMultiplier = modelMultiplier
			ruleKey = buildVIPModelRuleKey(rule.RuleKey, model)
		}
		break
	}
	if selectedMultiplier <= 0 {
		selectedMultiplier = 1
	}
	decision.Applied = selectedMultiplier != 1
	decision.BaseMultiplier = selectedMultiplier
	decision.RuleKey = ruleKey
	decision.OriginalCost = cost.ActualCost
	decision.FinalMultiplier = baseMultiplier * selectedMultiplier
	cost.ActualCost = cost.TotalCost * decision.FinalMultiplier
	decision.DiscountAmount = math.Max(0, decision.OriginalCost-cost.ActualCost)
	return decision, nil
}

func (s *VIPService) OnRechargeSuccess(ctx context.Context, userID int64, amount float64) {
	if s == nil || s.statsRepo == nil || amount <= 0 {
		return
	}
	if err := s.statsRepo.AddRecharge(ctx, userID, amount); err != nil {
		return
	}
}

func (s *VIPService) QueueSpend(ctx context.Context, userID int64, amount float64) {
	if s == nil || s.statsRepo == nil || amount <= 0 {
		return
	}
	_ = s.statsRepo.AddPendingSpend(ctx, userID, amount)
}

func computeVIPProgress(current *VIPSummary, next *VIPNextLevel) float64 {
	if current == nil || next == nil {
		return 0
	}
	progresses := make([]float64, 0, 2)
	if next.RequiredRecharge > 0 {
		progresses = append(progresses, math.Min(100, current.RechargeTotal/next.RequiredRecharge*100))
	}
	if next.RequiredSpend > 0 {
		progresses = append(progresses, math.Min(100, current.SpendTotal/next.RequiredSpend*100))
	}
	if len(progresses) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range progresses {
		sum += v
	}
	return sum / float64(len(progresses))
}

func buildVIPUnlockConditionLabel(rule VIPRule) string {
	parts := make([]string, 0, 2)
	if rule.RequiredRecharge > 0 {
		parts = append(parts, fmt.Sprintf("实付满 %.2f 元", rule.RequiredRecharge))
	}
	if rule.RequiredSpend > 0 {
		parts = append(parts, fmt.Sprintf("累计消费满 %.2f", rule.RequiredSpend))
	}
	if len(parts) == 0 {
		return "无门槛"
	}
	return strings.Join(parts, " 且 ")
}

func resolveVIPModelMultiplier(modelRules map[string]float64, model string) float64 {
	if len(modelRules) == 0 || strings.TrimSpace(model) == "" {
		return 0
	}
	if value, ok := modelRules[model]; ok && value > 0 {
		return value
	}
	for pattern, value := range modelRules {
		if value <= 0 {
			continue
		}
		if matchModelPattern(pattern, model) {
			return value
		}
	}
	return 0
}

func buildVIPModelRuleKey(ruleKey string, model string) string {
	ruleKey = strings.TrimSpace(ruleKey)
	if ruleKey == "" {
		ruleKey = "vip"
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return ruleKey
	}
	return ruleKey + ":" + model
}
