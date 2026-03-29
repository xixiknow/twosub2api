package handler

import (
	"database/sql"
	"math"
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type ReferralHandler struct {
	userService *service.UserService
	settingSvc  *service.SettingService
	sqlDB       *sql.DB
}

func NewReferralHandler(userService *service.UserService, settingSvc *service.SettingService, sqlDB *sql.DB) *ReferralHandler {
	return &ReferralHandler{
		userService: userService,
		settingSvc:  settingSvc,
		sqlDB:       sqlDB,
	}
}

type ReferralInfoResponse struct {
	ReferralCode   string  `json:"referral_code"`
	CommissionRate float64 `json:"commission_rate"`
	TotalEarnings  float64 `json:"total_earnings"`
	TotalReferred  int     `json:"total_referred"`
}

// GetReferralInfo returns the current user's referral information
// GET /api/v1/user/referral
func (h *ReferralHandler) GetReferralInfo(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	code := ""
	if user.ReferralCode != nil {
		code = *user.ReferralCode
	}

	rate := h.settingSvc.GetReferralCommissionRate(c.Request.Context())

	var totalEarnings float64
	var totalReferred int

	if h.sqlDB != nil {
		_ = h.sqlDB.QueryRowContext(c.Request.Context(),
			`SELECT COALESCE(SUM(commission_amount), 0) FROM referral_commissions WHERE referrer_id = $1`,
			subject.UserID,
		).Scan(&totalEarnings)

		_ = h.sqlDB.QueryRowContext(c.Request.Context(),
			`SELECT COUNT(DISTINCT id) FROM users WHERE referrer_id = $1 AND deleted_at IS NULL`,
			subject.UserID,
		).Scan(&totalReferred)
	}

	response.Success(c, ReferralInfoResponse{
		ReferralCode:   code,
		CommissionRate: rate,
		TotalEarnings:  totalEarnings,
		TotalReferred:  totalReferred,
	})
}

type CommissionRecord struct {
	ID               int64   `json:"id"`
	ReferredUserID   int64   `json:"referred_user_id"`
	OrderAmount      float64 `json:"order_amount"`
	CommissionRate   float64 `json:"commission_rate"`
	CommissionAmount float64 `json:"commission_amount"`
	CreatedAt        string  `json:"created_at"`
}

type CommissionsResponse struct {
	Items    []CommissionRecord `json:"items"`
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Pages    int                `json:"pages"`
}

// ReferredUserRecord represents a user referred by the current user
type ReferredUserRecord struct {
	Email           string  `json:"email"`
	CreatedAt       string  `json:"created_at"`
	TotalCommission float64 `json:"total_commission"`
}

// ReferredUsersResponse is the paginated response for referred users
type ReferredUsersResponse struct {
	Items    []ReferredUserRecord `json:"items"`
	Total    int                  `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
	Pages    int                  `json:"pages"`
}

// GetReferredUsers returns paginated list of users referred by the current user
// GET /api/v1/user/referral/users
func (h *ReferralHandler) GetReferredUsers(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int
	if h.sqlDB != nil {
		_ = h.sqlDB.QueryRowContext(c.Request.Context(),
			`SELECT COUNT(*) FROM users WHERE referrer_id = $1 AND deleted_at IS NULL`,
			subject.UserID,
		).Scan(&total)
	}

	pages := 0
	if total > 0 {
		pages = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	items := make([]ReferredUserRecord, 0)
	if h.sqlDB != nil && total > 0 {
		offset := (page - 1) * pageSize
		rows, err := h.sqlDB.QueryContext(c.Request.Context(),
			`SELECT u.email, u.created_at, COALESCE(rc.total, 0) AS total_commission
			 FROM users u
			 LEFT JOIN (
			   SELECT referred_user_id, SUM(commission_amount) AS total
			   FROM referral_commissions
			   WHERE referrer_id = $1
			   GROUP BY referred_user_id
			 ) rc ON rc.referred_user_id = u.id
			 WHERE u.referrer_id = $1 AND u.deleted_at IS NULL
			 ORDER BY u.created_at DESC
			 LIMIT $2 OFFSET $3`,
			subject.UserID, pageSize, offset,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var r ReferredUserRecord
				var email string
				var createdAt time.Time
				var totalComm float64
				if err := rows.Scan(&email, &createdAt, &totalComm); err == nil {
					// Mask email: show first 2 chars + *** + @domain
					r.Email = email
					for i, ch := range email {
						if ch == '@' {
							if i > 2 {
								r.Email = email[:2] + "***" + email[i:]
							} else if i > 0 {
								r.Email = email[:1] + "***" + email[i:]
							}
							break
						}
					}
					r.CreatedAt = createdAt.Format(time.RFC3339)
					r.TotalCommission = totalComm
					items = append(items, r)
				}
			}
		}
	}

	response.Success(c, ReferredUsersResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
	})
}

// GetCommissions returns paginated commission records for the current user
// GET /api/v1/user/referral/commissions
func (h *ReferralHandler) GetCommissions(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int
	if h.sqlDB != nil {
		_ = h.sqlDB.QueryRowContext(c.Request.Context(),
			`SELECT COUNT(*) FROM referral_commissions WHERE referrer_id = $1`,
			subject.UserID,
		).Scan(&total)
	}

	pages := 0
	if total > 0 {
		pages = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	items := make([]CommissionRecord, 0)
	if h.sqlDB != nil && total > 0 {
		offset := (page - 1) * pageSize
		rows, err := h.sqlDB.QueryContext(c.Request.Context(),
			`SELECT id, referred_user_id, order_amount, commission_rate, commission_amount, created_at
			 FROM referral_commissions
			 WHERE referrer_id = $1
			 ORDER BY id DESC
			 LIMIT $2 OFFSET $3`,
			subject.UserID, pageSize, offset,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var r CommissionRecord
				var createdAt time.Time
				if err := rows.Scan(&r.ID, &r.ReferredUserID, &r.OrderAmount, &r.CommissionRate, &r.CommissionAmount, &createdAt); err == nil {
					r.CreatedAt = createdAt.Format(time.RFC3339)
					items = append(items, r)
				}
			}
		}
	}

	response.Success(c, CommissionsResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
	})
}
