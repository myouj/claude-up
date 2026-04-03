package service

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"

	"prompt-vault/models"
)

// EvalService handles evaluation set operations
type EvalService struct {
	db             *gorm.DB
	scoringService *ScoringService
}

// NewEvalService creates a new EvalService
func NewEvalService(db *gorm.DB) *EvalService {
	return &EvalService{
		db:             db,
		scoringService: NewScoringService(),
	}
}

// CreateEvalSet creates a new evaluation set for a prompt
func (s *EvalService) CreateEvalSet(promptID uint, name string, cases []models.EvalCase, weights models.EvalWeights) (*models.EvalSet, error) {
	// Verify prompt exists
	var prompt models.Prompt
	if err := s.db.First(&prompt, promptID).Error; err != nil {
		return nil, errors.New("prompt not found")
	}

	casesJSON, err := json.Marshal(cases)
	if err != nil {
		return nil, err
	}

	weightsJSON, err := json.Marshal(weights)
	if err != nil {
		return nil, err
	}

	evalSet := &models.EvalSet{
		PromptID: promptID,
		Name:     name,
		Cases:    string(casesJSON),
		Weights:  string(weightsJSON),
	}

	if err := s.db.Create(evalSet).Error; err != nil {
		return nil, err
	}

	return evalSet, nil
}

// GetEvalSet retrieves an evaluation set by ID
func (s *EvalService) GetEvalSet(id uint) (*models.EvalSet, error) {
	var evalSet models.EvalSet
	if err := s.db.First(&evalSet, id).Error; err != nil {
		return nil, errors.New("eval set not found")
	}
	return &evalSet, nil
}

// ListEvalSetsByPrompt returns all evaluation sets for a prompt
func (s *EvalService) ListEvalSetsByPrompt(promptID uint) ([]models.EvalSet, error) {
	var evalSets []models.EvalSet
	if err := s.db.Where("prompt_id = ?", promptID).Order("created_at DESC").Find(&evalSets).Error; err != nil {
		return nil, err
	}
	return evalSets, nil
}

// UpdateEvalSet updates an evaluation set
func (s *EvalService) UpdateEvalSet(id uint, name string, cases []models.EvalCase, weights models.EvalWeights) (*models.EvalSet, error) {
	var evalSet models.EvalSet
	if err := s.db.First(&evalSet, id).Error; err != nil {
		return nil, errors.New("eval set not found")
	}

	casesJSON, err := json.Marshal(cases)
	if err != nil {
		return nil, err
	}

	weightsJSON, err := json.Marshal(weights)
	if err != nil {
		return nil, err
	}

	evalSet.Name = name
	evalSet.Cases = string(casesJSON)
	evalSet.Weights = string(weightsJSON)

	if err := s.db.Save(&evalSet).Error; err != nil {
		return nil, err
	}

	return &evalSet, nil
}

// DeleteEvalSet deletes an evaluation set
func (s *EvalService) DeleteEvalSet(id uint) error {
	result := s.db.Delete(&models.EvalSet{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("eval set not found")
	}
	return nil
}

// GenerateAutoEvalSet generates an evaluation set with auto-generated test cases
func (s *EvalService) GenerateAutoEvalSet(promptID uint, name string, caseCount int) (*models.EvalSet, error) {
	// Get the prompt
	var prompt models.Prompt
	if err := s.db.First(&prompt, promptID).Error; err != nil {
		return nil, errors.New("prompt not found")
	}

	// Generate test cases
	cases, err := s.scoringService.GenerateEvalCases(&prompt, caseCount)
	if err != nil {
		return nil, err
	}

	// Use default weights
	weights := s.scoringService.GetDefaultWeights()

	return s.CreateEvalSet(promptID, name, cases, weights)
}

// RunEval runs evaluation on a prompt using the specified eval set
func (s *EvalService) RunEval(promptID uint, evalSetID uint) (*ScoreResult, error) {
	// Get the prompt
	var prompt models.Prompt
	if err := s.db.First(&prompt, promptID).Error; err != nil {
		return nil, errors.New("prompt not found")
	}

	// Get the eval set
	evalSet, err := s.GetEvalSet(evalSetID)
	if err != nil {
		return nil, err
	}

	// Run scoring
	result := s.scoringService.Score(&prompt)

	// Parse weights to potentially adjust scoring
	var weights models.EvalWeights
	if evalSet.Weights != "" {
		json.Unmarshal([]byte(evalSet.Weights), &weights)
	}

	// Recalculate with custom weights if provided
	if weights.Clarity > 0 || weights.Completeness > 0 || weights.Example > 0 || weights.Role > 0 {
		scores := map[string]float64{
			"clarity":      result.Clarity,
			"completeness": result.Completeness,
			"example":      result.Example,
			"role":         result.Role,
		}
		result.Overall = s.scoringService.CalculateWeightedScore(weights, scores)
	}

	return result, nil
}

// ValidateEvalSetCases validates that eval set cases are properly formatted
func (s *EvalService) ValidateEvalSetCases(cases []models.EvalCase) error {
	if len(cases) < 5 {
		return errors.New("eval set must have at least 5 cases")
	}
	if len(cases) > 20 {
		return errors.New("eval set must have at most 20 cases")
	}
	return nil
}
