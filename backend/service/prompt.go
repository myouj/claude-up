package service

import (
	"prompt-vault/models"

	"gorm.io/gorm"
)

// PromptService encapsulates business logic for Prompt entities.
type PromptService struct {
	db *gorm.DB
}

// NewPromptService creates a new PromptService.
func NewPromptService(db *gorm.DB) *PromptService {
	return &PromptService{db: db}
}

// EnsureVersion creates a new version if the content has changed.
// Returns whether a new version was created and the version number.
// Returns gorm.ErrRecordNotFound if the prompt does not exist.
func (s *PromptService) EnsureVersion(promptID uint, newContent, comment string) (created bool, versionNum int, err error) {
	// Verify prompt exists first.
	var prompt models.Prompt
	if err := s.db.First(&prompt, promptID).Error; err != nil {
		return false, 0, err
	}

	var latest models.PromptVersion
	if err := s.db.Where("prompt_id = ?", promptID).
		Order("version DESC").First(&latest).Error; err != nil && err != gorm.ErrRecordNotFound {
		return false, 0, err
	}
	_ = prompt

	// Only create a version if content differs from latest
	if latest.Content == newContent {
		return false, latest.Version, nil
	}

	newVersion := latest.Version + 1
	if latest.ID == 0 {
		newVersion = 1
	}

	version := models.PromptVersion{
		PromptID: promptID,
		Version:  newVersion,
		Content:  newContent,
		Comment: comment,
	}
	if err := s.db.Create(&version).Error; err != nil {
		return false, 0, err
	}

	// Also update the prompt's current content
	s.db.Model(&models.Prompt{}).Where("id = ?", promptID).
		Update("content", newContent)

	return true, newVersion, nil
}

// DeleteWithVersionsAndTests deletes a prompt along with its versions and test records.
// Returns gorm.ErrRecordNotFound if the prompt does not exist.
func (s *PromptService) DeleteWithVersionsAndTests(promptID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Verify prompt exists within the transaction.
		var prompt models.Prompt
		if err := tx.First(&prompt, promptID).Error; err != nil {
			return err
		}
		_ = prompt

		if err := tx.Where("prompt_id = ?", promptID).Delete(&models.PromptVersion{}).Error; err != nil {
			return err
		}
		if err := tx.Where("prompt_id = ?", promptID).Delete(&models.TestRecord{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Prompt{}, promptID).Error
	})
}

// CountVersions returns the number of versions for a prompt.
func (s *PromptService) CountVersions(promptID uint) (int64, error) {
	var count int64
	err := s.db.Model(&models.PromptVersion{}).Where("prompt_id = ?", promptID).Count(&count).Error
	return count, err
}
