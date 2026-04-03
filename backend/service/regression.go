package service

import (
	"time"

	"gorm.io/gorm"

	"prompt-vault/models"
)

type RegressionService struct {
	db         *gorm.DB
	scoringSvc *ScoringService
}

func NewRegressionService(db *gorm.DB, scoringSvc *ScoringService) *RegressionService {
	return &RegressionService{db: db, scoringSvc: scoringSvc}
}

type RegressionReport struct {
	PromptID      uint      `json:"prompt_id"`
	PromptTitle   string    `json:"prompt_title"`
	OldModel      string    `json:"old_model"`
	NewModel      string    `json:"new_model"`
	OldScore      float64   `json:"old_score"`
	NewScore      float64   `json:"new_score"`
	ScoreDelta    float64   `json:"score_delta"`
	HasRegression bool      `json:"has_regression"`
	GeneratedAt   time.Time `json:"generated_at"`
}

func (s *RegressionService) Detect(promptID uint, oldModel, newModel string) (*RegressionReport, error) {
	var prompt models.Prompt
	if err := s.db.First(&prompt, promptID).Error; err != nil {
		return nil, err
	}

	oldScore := s.scoringSvc.Score(&prompt)

	newScore := s.scoringSvc.Score(&prompt)

	report := &RegressionReport{
		PromptID:      promptID,
		PromptTitle:   prompt.Title,
		OldModel:      oldModel,
		NewModel:      newModel,
		OldScore:      oldScore.Overall,
		NewScore:      newScore.Overall,
		ScoreDelta:    newScore.Overall - oldScore.Overall,
		HasRegression: newScore.Overall < oldScore.Overall,
		GeneratedAt:   time.Now(),
	}

	return report, nil
}

func (s *RegressionService) GetReport(promptID uint, oldModel, newModel string) (*RegressionReport, error) {
	return s.Detect(promptID, oldModel, newModel)
}
