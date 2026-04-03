package service

import (
	"encoding/json"
	"math"
	"math/rand"

	"gorm.io/gorm"

	"prompt-vault/models"
)

// ABTestService encapsulates business logic for ABTest entities.
type ABTestService struct {
	db *gorm.DB
}

// NewABTestService creates a new ABTestService.
func NewABTestService(db *gorm.DB) *ABTestService {
	return &ABTestService{db: db}
}

// Create creates a new A/B test.
func (s *ABTestService) Create(promptID uint, name, configJSON string) (*models.ABTest, error) {
	// Validate config JSON
	var config models.ABTestConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, err
	}

	// Set defaults
	if config.Alpha == 0 {
		config.Alpha = 0.05
	}
	if config.MinRuns == 0 {
		config.MinRuns = 10
	}
	if config.MaxRuns == 0 {
		config.MaxRuns = 100
	}

	// Re-marshal with defaults
	configJSONBytes, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	abTest := &models.ABTest{
		PromptID: promptID,
		Name:     name,
		Config:   string(configJSONBytes),
		Status:   models.ABTestStatusPending,
	}

	if err := s.db.Create(abTest).Error; err != nil {
		return nil, err
	}

	return abTest, nil
}

// GetByID retrieves an ABTest by ID.
func (s *ABTestService) GetByID(id uint) (*models.ABTest, error) {
	var abTest models.ABTest
	if err := s.db.First(&abTest, id).Error; err != nil {
		return nil, err
	}
	return &abTest, nil
}

// UpdateStatus updates the status of an A/B test.
func (s *ABTestService) UpdateStatus(id uint, status string) error {
	return s.db.Model(&models.ABTest{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateResult updates the result JSON of an A/B test.
func (s *ABTestService) UpdateResult(id uint, resultJSON string) error {
	return s.db.Model(&models.ABTest{}).Where("id = ?", id).Update("result", resultJSON).Error
}

// RecordResult records a single result for an A/B test run.
func (s *ABTestService) RecordResult(abTestID uint, runIndex int, variant string, score float64, latencyMs int) (*models.ABTestResult, error) {
	result := &models.ABTestResult{
		ABTestID:  abTestID,
		RunIndex:  runIndex,
		Variant:   variant,
		Score:     score,
		LatencyMs: latencyMs,
	}

	if err := s.db.Create(result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

// GetResults retrieves all results for an A/B test.
func (s *ABTestService) GetResults(abTestID uint) ([]models.ABTestResult, error) {
	var results []models.ABTestResult
	if err := s.db.Where("ab_test_id = ?", abTestID).Order("run_index ASC, id ASC").Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

// GetResultsSummary returns summary statistics for an A/B test.
func (s *ABTestService) GetResultsSummary(abTestID uint) (*models.SPRTResult, error) {
	results, err := s.GetResults(abTestID)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return &models.SPRTResult{}, nil
	}

	var sumScoreA, sumScoreB float64
	var sumLatencyA, sumLatencyB float64
	var countA, countB int

	for _, r := range results {
		if r.Variant == "A" {
			sumScoreA += r.Score
			sumLatencyA += float64(r.LatencyMs)
			countA++
		} else {
			sumScoreB += r.Score
			sumLatencyB += float64(r.LatencyMs)
			countB++
		}
	}

	avgScoreA := 0.0
	avgScoreB := 0.0
	avgLatencyA := 0.0
	avgLatencyB := 0.0

	if countA > 0 {
		avgScoreA = sumScoreA / float64(countA)
		avgLatencyA = sumLatencyA / float64(countA)
	}
	if countB > 0 {
		avgScoreB = sumScoreB / float64(countB)
		avgLatencyB = sumLatencyB / float64(countB)
	}

	return &models.SPRTResult{
		TotalRunsA:      countA,
		TotalRunsB:      countB,
		AverageScoreA:   avgScoreA,
		AverageScoreB:   avgScoreB,
		AverageLatencyA: avgLatencyA,
		AverageLatencyB: avgLatencyB,
	}, nil
}

// List returns all A/B tests with pagination.
func (s *ABTestService) List(offset, limit int) ([]models.ABTest, int64, error) {
	var abTests []models.ABTest
	var total int64

	if err := s.db.Model(&models.ABTest{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&abTests).Error; err != nil {
		return nil, 0, err
	}

	return abTests, total, nil
}

// ListByPromptID returns all A/B tests for a specific prompt.
func (s *ABTestService) ListByPromptID(promptID uint, offset, limit int) ([]models.ABTest, int64, error) {
	var abTests []models.ABTest
	var total int64

	query := s.db.Model(&models.ABTest{}).Where("prompt_id = ?", promptID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&abTests).Error; err != nil {
		return nil, 0, err
	}

	return abTests, total, nil
}

// Delete deletes an A/B test and its results.
func (s *ABTestService) Delete(id uint) error {
	// First check if the test exists
	var count int64
	s.db.Model(&models.ABTest{}).Where("id = ?", id).Count(&count)
	if count == 0 {
		return gorm.ErrRecordNotFound
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete results first
		if err := tx.Where("ab_test_id = ?", id).Delete(&models.ABTestResult{}).Error; err != nil {
			return err
		}
		// Delete the test
		return tx.Delete(&models.ABTest{}, id).Error
	})
}

// CheckSignificance performs SPRT (Sequential Probability Ratio Test) analysis.
// It returns the SPRT result with whether a winner has been determined.
func (s *ABTestService) CheckSignificance(abTestID uint) (*models.SPRTResult, error) {
	results, err := s.GetResults(abTestID)
	if err != nil {
		return nil, err
	}

	abTest, err := s.GetByID(abTestID)
	if err != nil {
		return nil, err
	}

	// Parse config
	var config models.ABTestConfig
	if err := json.Unmarshal([]byte(abTest.Config), &config); err != nil {
		return nil, err
	}

	// Get summary statistics
	summary, err := s.GetResultsSummary(abTestID)
	if err != nil {
		return nil, err
	}

	// Calculate SPRT
	alpha := config.Alpha
	if alpha == 0 {
		alpha = 0.05
	}

	// Calculate log-likelihood ratio using mock evaluation
	// For SPRT, we compare the probability of each variant being better
	logLambdaA := 0.0
	logLambdaB := 0.0

	for _, r := range results {
		scoreDiff := r.Score // In real impl, this would be compared against a baseline
		if r.Variant == "A" {
			logLambdaA += math.Log(1 + scoreDiff)
		} else {
			logLambdaB += math.Log(1 + scoreDiff)
		}
	}

	// SPRT boundaries
	logEpsilon := math.Log(alpha / (1 - alpha))
	lowerBound := math.Log((1-alpha)/alpha) + logEpsilon
	upperBound := math.Log(alpha/(1-alpha)) - logEpsilon
	_ = lowerBound // Reserved for future use when tracking both directions

	// Determine significance
	isSignificant := false
	winner := ""

	totalRuns := summary.TotalRunsA + summary.TotalRunsB
	if totalRuns >= config.MinRuns {
		if logLambdaA > upperBound {
			isSignificant = true
			winner = "A"
		} else if logLambdaB > upperBound {
			isSignificant = true
			winner = "B"
		}
	}

	// Calculate probabilities (normalized)
	total := logLambdaA + logLambdaB
	if total == 0 {
		total = 1 // Avoid division by zero
	}

	pA := math.Exp(logLambdaA) / (math.Exp(logLambdaA) + math.Exp(logLambdaB))
	pB := math.Exp(logLambdaB) / (math.Exp(logLambdaA) + math.Exp(logLambdaB))

	return &models.SPRTResult{
		Winner:          winner,
		IsSignificant:   isSignificant,
		PA:              pA,
		PB:              pB,
		TotalRunsA:      summary.TotalRunsA,
		TotalRunsB:      summary.TotalRunsB,
		LambdaA:         math.Exp(logLambdaA),
		LambdaB:         math.Exp(logLambdaB),
		AverageScoreA:   summary.AverageScoreA,
		AverageScoreB:   summary.AverageScoreB,
		AverageLatencyA: summary.AverageLatencyA,
		AverageLatencyB: summary.AverageLatencyB,
	}, nil
}

// RunIteration simulates a single iteration of the A/B test by generating mock results.
// In production, this would call the actual AI provider.
func (s *ABTestService) RunIteration(abTestID uint) (*models.ABTestResult, error) {
	abTest, err := s.GetByID(abTestID)
	if err != nil {
		return nil, err
	}

	// Parse config
	var config models.ABTestConfig
	if err := json.Unmarshal([]byte(abTest.Config), &config); err != nil {
		return nil, err
	}

	// Get current run count
	results, err := s.GetResults(abTestID)
	if err != nil {
		return nil, err
	}

	runIndex := len(results) + 1

	// Determine which variant to run (alternating for fairness)
	variant := "A"
	if runIndex%2 == 0 {
		variant = "B"
	}

	// Simulate mock score based on variant
	// In real implementation, this would call the AI provider
	baseScore := 0.5 + rand.Float64()*0.3 // Base score between 0.5 and 0.8
	var latencyMs int

	if variant == "A" {
		latencyMs = 100 + rand.Intn(200) // 100-300ms
	} else {
		latencyMs = 120 + rand.Intn(250) // 120-370ms
		baseScore *= (0.95 + rand.Float64()*0.1) // Slightly different for B
	}

	score := baseScore

	result, err := s.RecordResult(abTestID, runIndex, variant, score, latencyMs)
	if err != nil {
		return nil, err
	}

	// Check significance after recording
	sprtResult, err := s.CheckSignificance(abTestID)
	if err != nil {
		return result, nil // Return result even if significance check fails
	}

	// Update result in database
	resultJSON, _ := json.Marshal(sprtResult)
	s.UpdateResult(abTestID, string(resultJSON))

	// Check if test should be marked as completed
	if sprtResult.IsSignificant && config.EarlyStop {
		s.UpdateStatus(abTestID, models.ABTestStatusCompleted)
	} else if runIndex >= config.MaxRuns {
		s.UpdateStatus(abTestID, models.ABTestStatusCompleted)
	}

	return result, nil
}
