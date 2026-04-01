package service

import (
	"fmt"

	"prompt-vault/models"

	"gorm.io/gorm"
)

// SkillService encapsulates business logic for Skill entities.
type SkillService struct {
	db *gorm.DB
}

// NewSkillService creates a new SkillService.
func NewSkillService(db *gorm.DB) *SkillService {
	return &SkillService{db: db}
}

// Clone creates a copy of a skill with "(Copy)" appended to the name.
// The cloned skill always has source set to "custom".
func (s *SkillService) Clone(skillID uint) (*models.Skill, error) {
	var skill models.Skill
	if err := s.db.First(&skill, skillID).Error; err != nil {
		return nil, err
	}

	clone := models.Skill{
		Name:        skill.Name + " (Copy)",
		Description: skill.Description,
		Content:     skill.Content,
		ContentCN:   skill.ContentCN,
		Category:    skill.Category,
		Source:      "custom",
	}

	if err := s.db.Create(&clone).Error; err != nil {
		return nil, err
	}
	return &clone, nil
}

// Delete removes a skill by ID. Returns gorm.ErrRecordNotFound if not found.
func (s *SkillService) Delete(skillID uint) error {
	result := s.db.Delete(&models.Skill{}, skillID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Count returns the total number of skills.
func (s *SkillService) Count() (int64, error) {
	var count int64
	err := s.db.Model(&models.Skill{}).Count(&count).Error
	return count, err
}

// GetByID retrieves a skill by ID. Returns gorm.ErrRecordNotFound if not found.
func (s *SkillService) GetByID(skillID uint) (*models.Skill, error) {
	var skill models.Skill
	if err := s.db.First(&skill, skillID).Error; err != nil {
		return nil, err
	}
	return &skill, nil
}

// BatchClone creates copies of multiple skills in a single transaction.
// Returns the number successfully cloned and any errors encountered.
func (s *SkillService) BatchClone(skillIDs []uint) ([]models.Skill, []BatchError, error) {
	var cloned []models.Skill
	var batchErrs []BatchError

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, id := range skillIDs {
			var skill models.Skill
			if err := tx.First(&skill, id).Error; err != nil {
				batchErrs = append(batchErrs, BatchError{ID: id, Err: err.Error()})
				continue
			}

			clone := models.Skill{
				Name:        skill.Name + " (Copy)",
				Description: skill.Description,
				Content:     skill.Content,
				ContentCN:   skill.ContentCN,
				Category:    skill.Category,
				Source:      "custom",
			}
			if err := tx.Create(&clone).Error; err != nil {
				batchErrs = append(batchErrs, BatchError{ID: id, Err: err.Error()})
				continue
			}
			cloned = append(cloned, clone)
		}
		return nil
	})

	return cloned, batchErrs, err
}

// BatchError records an error for a specific entity ID during batch operations.
type BatchError struct {
	ID  uint
	Err string
}

// CloneWithActivity wraps Clone and returns a formatted details JSON string.
func (s *SkillService) CloneWithActivity(skillID uint) (*models.Skill, string, error) {
	clone, err := s.Clone(skillID)
	if err != nil {
		return nil, "", err
	}
	return clone, fmt.Sprintf(`{"from_id": %d}`, skillID), nil
}
