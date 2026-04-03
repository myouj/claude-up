package service

import (
	"errors"
	"time"

	"prompt-vault/models"

	"gorm.io/gorm"
)

// ErrQuotaExceeded is returned when quota is exceeded.
var ErrQuotaExceeded = errors.New("quota exceeded")

// ErrQuotaNotFound is returned when no quota exists for a provider.
var ErrQuotaNotFound = errors.New("quota not found")

// QuotaService handles quota checking and consumption.
type QuotaService struct {
	db *gorm.DB
}

// NewQuotaService creates a new QuotaService.
func NewQuotaService(db *gorm.DB) *QuotaService {
	return &QuotaService{db: db}
}

// GetQuota retrieves the quota for a provider (optionally filtered by model).
// If no quota exists, returns ErrQuotaNotFound.
func (s *QuotaService) GetQuota(provider string, model string) (*models.Quota, error) {
	var quota models.Quota
	query := s.db.Where("provider = ?", provider)
	if model != "" {
		query = query.Where("model = ?", model)
	}
	// Prefer model-specific quota, fall back to provider-level quota
	if model != "" {
		err := query.Where("model = ? OR model = ''", model).Order("model DESC").First(&quota).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrQuotaNotFound
			}
			return nil, err
		}
		return &quota, nil
	}
	err := query.First(&quota).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuotaNotFound
		}
		return nil, err
	}
	return &quota, nil
}

// Check verifies if the provider has sufficient quota for the given cost.
// Returns true if quota is sufficient, false if not, or error if quota not found.
// If reset time has passed, it will reset the usage in the database.
func (s *QuotaService) Check(provider string, cost int) (bool, error) {
	quota, err := s.GetQuota(provider, "")
	if err != nil {
		if errors.Is(err, ErrQuotaNotFound) {
			// No quota configured means unlimited
			return true, nil
		}
		return false, err
	}

	// Check if reset is needed and persist it
	if time.Now().After(quota.ResetAt) {
		quota.Usage = 0
		quota.ResetAt = nextMonth()
		if err := s.db.Save(&quota).Error; err != nil {
			return false, err
		}
		return true, nil
	}

	return quota.Usage+cost <= quota.Limit, nil
}

// Consume deducts the specified cost from the provider's quota.
// Returns ErrQuotaExceeded if the cost would exceed the limit.
// Uses a transaction to prevent race conditions.
func (s *QuotaService) Consume(provider string, cost int) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var quota models.Quota
		if err := tx.Where("provider = ?", provider).First(&quota).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// No quota configured means unlimited
				return nil
			}
			return err
		}

		// Check if reset is needed
		if time.Now().After(quota.ResetAt) {
			// Reset usage
			quota.Usage = 0
			quota.ResetAt = nextMonth()
			return tx.Save(&quota).Error
		}

		// Check if consumption would exceed limit
		if quota.Usage+cost > quota.Limit {
			return ErrQuotaExceeded
		}

		// Atomic update: increment usage with WHERE clause to prevent race condition
		result := tx.Model(&models.Quota{}).
			Where("provider = ? AND id = ? AND usage + ? <= ?", provider, quota.ID, cost, quota.Limit).
			Update("usage", gorm.Expr("usage + ?", cost))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			// Race condition detected or limit changed
			return ErrQuotaExceeded
		}
		return nil
	})
}

// GetUsage returns the current usage for a provider.
func (s *QuotaService) GetUsage(provider string) (int, error) {
	quota, err := s.GetQuota(provider, "")
	if err != nil {
		if errors.Is(err, ErrQuotaNotFound) {
			return 0, nil
		}
		return 0, err
	}

	// Check if reset is needed
	if time.Now().After(quota.ResetAt) {
		return 0, nil
	}

	return quota.Usage, nil
}

// CreateOrUpdate creates or updates a quota entry for a provider.
func (s *QuotaService) CreateOrUpdate(provider string, model string, limit int) (*models.Quota, error) {
	var quota models.Quota
	err := s.db.Where("provider = ? AND model = ?", provider, model).First(&quota).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if quota.ID != 0 {
		// Update existing
		quota.Limit = limit
		if quota.ResetAt.IsZero() {
			quota.ResetAt = nextMonth()
		}
	} else {
		// Create new
		quota = models.Quota{
			Provider: provider,
			Model:    model,
			Limit:    limit,
			Usage:    0,
			ResetAt:  nextMonth(),
		}
	}

	if err := s.db.Save(&quota).Error; err != nil {
		return nil, err
	}
	return &quota, nil
}

// ResetUsage resets the usage for a provider to 0.
func (s *QuotaService) ResetUsage(provider string) error {
	return s.db.Model(&models.Quota{}).Where("provider = ?", provider).Updates(map[string]interface{}{
		"usage":    0,
		"reset_at": nextMonth(),
	}).Error
}

// nextMonth returns the first day of the next month.
// Uses AddDate to properly handle December -> January transition.
func nextMonth() time.Time {
	now := time.Now()
	// Add one month (AddDate handles year rollover properly)
	nextMonth := now.AddDate(0, 1, 0)
	// Return first day of next month at midnight
	return time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, now.Location())
}
