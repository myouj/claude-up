package service

import (
	"testing"

	"prompt-vault/models"
)

func TestScoringService_ScoreClarity(t *testing.T) {
	svc := NewScoringService()

	tests := []struct {
		name    string
		content string
		wantMin float64
		wantMax float64
	}{
		{
			name:    "good prompt with variables",
			content: "You are a helpful assistant. {{user_name}} please help with {{task}}. ## Output Format: JSON",
			wantMin: 50,
			wantMax: 100,
		},
		{
			name:    "short prompt without variables",
			content: "Help me",
			wantMin: 30,
			wantMax: 70,
		},
		{
			name:    "prompt with properly formatted variables",
			content: "Hello {{name}}, your email is {{email}}. Please confirm.",
			wantMin: 50,
			wantMax: 100,
		},
		{
			name:    "prompt with poorly formatted variables",
			content: "Hello {{ name }}, your email is {{ email }}. Please confirm.",
			wantMin: 40,
			wantMax: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.ScoreClarity(tt.content)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("ScoreClarity() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestScoringService_ScoreCompleteness(t *testing.T) {
	svc := NewScoringService()

	prompt := &models.Prompt{
		Content: "You are a helpful assistant. Task: Help users. Output: JSON format. Requirements: Be concise. Context: User needs assistance.",
	}

	got := svc.ScoreCompleteness(prompt)
	if got < 50 || got > 100 {
		t.Errorf("ScoreCompleteness() = %v, want between 50 and 100", got)
	}
}

func TestScoringService_ScoreExample(t *testing.T) {
	svc := NewScoringService()

	prompt := &models.Prompt{
		Content: "You are a developer. Example: const x = 1; Example 2: const y = 2;",
	}

	got := svc.ScoreExample(prompt)
	if got < 50 || got > 100 {
		t.Errorf("ScoreExample() = %v, want between 50 and 100", got)
	}
}

func TestScoringService_ScoreRole(t *testing.T) {
	svc := NewScoringService()

	tests := []struct {
		name    string
		content string
		wantMin float64
		wantMax float64
	}{
		{
			name:    "with role definition",
			content: "You are an expert developer with 10 years of experience. You are capable of...",
			wantMin: 70,
			wantMax: 100,
		},
		{
			name:    "without role definition",
			content: "Help me with my task",
			wantMin: 20,
			wantMax: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.ScoreRole(tt.content)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("ScoreRole() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestScoringService_Score(t *testing.T) {
	svc := NewScoringService()

	prompt := &models.Prompt{
		Content: "You are an expert software developer. Task: Help users write code. Output: Provide working code examples. Requirements: Follow best practices. Context: User is a beginner.",
	}

	result := svc.Score(prompt)

	if result.Clarity < 0 || result.Clarity > 100 {
		t.Errorf("Clarity score out of range: %v", result.Clarity)
	}
	if result.Completeness < 0 || result.Completeness > 100 {
		t.Errorf("Completeness score out of range: %v", result.Completeness)
	}
	if result.Example < 0 || result.Example > 100 {
		t.Errorf("Example score out of range: %v", result.Example)
	}
	if result.Role < 0 || result.Role > 100 {
		t.Errorf("Role score out of range: %v", result.Role)
	}
	if result.Overall < 0 || result.Overall > 100 {
		t.Errorf("Overall score out of range: %v", result.Overall)
	}
}

func TestScoringService_GenerateEvalCases(t *testing.T) {
	svc := NewScoringService()

	prompt := &models.Prompt{
		Content: "You are a helpful assistant. {{name}} please help with {{task}}.",
	}

	// Test default count (10)
	cases, err := svc.GenerateEvalCases(prompt, 10)
	if err != nil {
		t.Fatalf("GenerateEvalCases() error = %v", err)
	}
	if len(cases) != 10 {
		t.Errorf("GenerateEvalCases() returned %d cases, want 10", len(cases))
	}

	// Test minimum (5)
	cases, err = svc.GenerateEvalCases(prompt, 5)
	if err != nil {
		t.Fatalf("GenerateEvalCases() error = %v", err)
	}
	if len(cases) != 5 {
		t.Errorf("GenerateEvalCases() returned %d cases, want 5", len(cases))
	}

	// Test maximum (20)
	cases, err = svc.GenerateEvalCases(prompt, 20)
	if err != nil {
		t.Fatalf("GenerateEvalCases() error = %v", err)
	}
	if len(cases) != 20 {
		t.Errorf("GenerateEvalCases() returned %d cases, want 20", len(cases))
	}

	// Test clamping below minimum
	cases, err = svc.GenerateEvalCases(prompt, 3)
	if err != nil {
		t.Fatalf("GenerateEvalCases() error = %v", err)
	}
	if len(cases) != 5 {
		t.Errorf("GenerateEvalCases() returned %d cases, want 5 (clamped from 3)", len(cases))
	}

	// Test clamping above maximum
	cases, err = svc.GenerateEvalCases(prompt, 25)
	if err != nil {
		t.Fatalf("GenerateEvalCases() error = %v", err)
	}
	if len(cases) != 20 {
		t.Errorf("GenerateEvalCases() returned %d cases, want 20 (clamped from 25)", len(cases))
	}
}

func TestScoringService_GetDefaultWeights(t *testing.T) {
	svc := NewScoringService()
	weights := svc.GetDefaultWeights()

	if weights.Clarity != 0.30 {
		t.Errorf("Clarity weight = %v, want 0.30", weights.Clarity)
	}
	if weights.Completeness != 0.30 {
		t.Errorf("Completeness weight = %v, want 0.30", weights.Completeness)
	}
	if weights.Example != 0.25 {
		t.Errorf("Example weight = %v, want 0.25", weights.Example)
	}
	if weights.Role != 0.15 {
		t.Errorf("Role weight = %v, want 0.15", weights.Role)
	}
}

func TestScoringService_CalculateWeightedScore(t *testing.T) {
	svc := NewScoringService()

	weights := models.EvalWeights{
		Clarity:      0.30,
		Completeness: 0.30,
		Example:      0.25,
		Role:         0.15,
	}

	scores := map[string]float64{
		"clarity":      80,
		"completeness": 90,
		"example":      70,
		"role":         85,
	}

	// Expected: 0.30*80 + 0.30*90 + 0.25*70 + 0.15*85 = 24 + 27 + 17.5 + 12.75 = 81.25
	got := svc.CalculateWeightedScore(weights, scores)
	expected := 81.25
	if got < expected-1 || got > expected+1 {
		t.Errorf("CalculateWeightedScore() = %v, want approximately %v", got, expected)
	}
}
