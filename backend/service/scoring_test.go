package service

import (
	"testing"
)

func TestScoringService_ScorePrompt(t *testing.T) {
	svc := NewScoringService()

	result, err := svc.ScorePrompt("Test prompt content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total < 0 || result.Total > 1 {
		t.Errorf("Total: got %f, expected between 0 and 1", result.Total)
	}
	if result.Clarity < 0 || result.Clarity > 1 {
		t.Errorf("Clarity: got %f, expected between 0 and 1", result.Clarity)
	}
	if result.Complexity < 0 || result.Complexity > 1 {
		t.Errorf("Complexity: got %f, expected between 0 and 1", result.Complexity)
	}
	if result.Quality < 0 || result.Quality > 1 {
		t.Errorf("Quality: got %f, expected between 0 and 1", result.Quality)
	}
}

func TestScoringService_ScoreResponse(t *testing.T) {
	svc := NewScoringService()

	result, err := svc.ScoreResponse("Test response content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total < 0 || result.Total > 1 {
		t.Errorf("Total: got %f, expected between 0 and 1", result.Total)
	}
	if result.Clarity < 0 || result.Clarity > 1 {
		t.Errorf("Clarity: got %f, expected between 0 and 1", result.Clarity)
	}
	if result.Complexity < 0 || result.Complexity > 1 {
		t.Errorf("Complexity: got %f, expected between 0 and 1", result.Complexity)
	}
	if result.Quality < 0 || result.Quality > 1 {
		t.Errorf("Quality: got %f, expected between 0 and 1", result.Quality)
	}
}

func TestScoringService_MultipleScores(t *testing.T) {
	svc := NewScoringService()

	result1, _ := svc.ScorePrompt("First prompt")
	result2, _ := svc.ScorePrompt("Second prompt")

	if result1.Total == result2.Total {
		t.Error("expected different scores for different prompts (random component)")
	}
}
