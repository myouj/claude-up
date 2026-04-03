package service

import "math/rand"

type ScoreResult struct {
	Total      float64
	Clarity    float64
	Complexity float64
	Quality    float64
}

type ScoringService struct{}

func NewScoringService() *ScoringService {
	return &ScoringService{}
}

func (s *ScoringService) ScorePrompt(content string) (*ScoreResult, error) {
	r := rand.Float64()
	return &ScoreResult{
		Total:      0.5 + r*0.5,
		Clarity:    0.5 + r*0.3,
		Complexity: 0.5 + r*0.3,
		Quality:    0.5 + r*0.3,
	}, nil
}

func (s *ScoringService) ScoreResponse(response string) (*ScoreResult, error) {
	r := rand.Float64()
	return &ScoreResult{
		Total:      0.5 + r*0.5,
		Clarity:    0.5 + r*0.3,
		Complexity: 0.5 + r*0.3,
		Quality:    0.5 + r*0.3,
	}, nil
}
