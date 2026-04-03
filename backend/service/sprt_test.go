package service

import (
	"math"
	"testing"
)

func TestNewSPRTEngineWithDefaults(t *testing.T) {
	engine := NewSPRTEngineWithDefaults()

	if engine.config.Alpha != 0.05 {
		t.Errorf("Expected Alpha=0.05, got %v", engine.config.Alpha)
	}
	if engine.config.Beta != 0.20 {
		t.Errorf("Expected Beta=0.20, got %v", engine.config.Beta)
	}
	if engine.config.MinSamples != 15 {
		t.Errorf("Expected MinSamples=15, got %v", engine.config.MinSamples)
	}
	if engine.config.MaxSamples != 50 {
		t.Errorf("Expected MaxSamples=50, got %v", engine.config.MaxSamples)
	}
	if engine.config.P0 != 0.5 {
		t.Errorf("Expected P0=0.5, got %v", engine.config.P0)
	}
	if engine.config.P1 != 0.6 {
		t.Errorf("Expected P1=0.6, got %v", engine.config.P1)
	}
}

func TestNewSPRTEngineCustomConfig(t *testing.T) {
	config := SPRTConfig{
		Alpha:     0.01,
		Beta:      0.10,
		MinSamples: 20,
		MaxSamples: 100,
		P0:        0.4,
		P1:        0.6,
	}
	engine := NewSPRTEngine(config)

	if engine.config.Alpha != 0.01 {
		t.Errorf("Expected Alpha=0.01, got %v", engine.config.Alpha)
	}
	if engine.config.Beta != 0.10 {
		t.Errorf("Expected Beta=0.10, got %v", engine.config.Beta)
	}
	if engine.config.MinSamples != 20 {
		t.Errorf("Expected MinSamples=20, got %v", engine.config.MinSamples)
	}
	if engine.config.MaxSamples != 100 {
		t.Errorf("Expected MaxSamples=100, got %v", engine.config.MaxSamples)
	}
}

func TestSPRTEngineEmptyScores(t *testing.T) {
	engine := NewSPRTEngineWithDefaults()

	result := engine.Test([]float64{}, []float64{})
	if result.Decision != Continue {
		t.Errorf("Expected Continue decision for empty scores, got %v", result.Decision)
	}
	if result.N != 0 {
		t.Errorf("Expected N=0 for empty scores, got %v", result.N)
	}
}

func TestSPRTEngineBelowMinSamples(t *testing.T) {
	engine := NewSPRTEngineWithDefaults()

	// Only 10 samples - below the default MinSamples of 15
	scoresA := []float64{0.8, 0.7, 0.6, 0.9, 0.75, 0.85, 0.7, 0.8, 0.65, 0.9}
	scoresB := []float64{0.6, 0.5, 0.4, 0.7, 0.55, 0.65, 0.5, 0.6, 0.45, 0.7}

	result := engine.Test(scoresA, scoresB)
	if result.Decision != Continue {
		t.Errorf("Expected Continue when below MinSamples, got %v", result.Decision)
	}
}

func TestSPRTEngineRejectH0(t *testing.T) {
	engine := NewSPRTEngineWithDefaults()

	// A consistently beats B (strong evidence for H1)
	scoresA := []float64{
		0.9, 0.85, 0.88, 0.92, 0.87,
		0.91, 0.86, 0.89, 0.90, 0.88,
		0.87, 0.92, 0.85, 0.91, 0.86,
		0.89, 0.90, 0.88, 0.87, 0.91,
	}
	scoresB := []float64{
		0.3, 0.25, 0.28, 0.32, 0.35,
		0.29, 0.26, 0.31, 0.27, 0.33,
		0.28, 0.34, 0.30, 0.26, 0.29,
		0.31, 0.28, 0.32, 0.27, 0.30,
	}

	result := engine.Test(scoresA, scoresB)
	if result.Decision != Reject {
		t.Errorf("Expected Reject (A is better), got %v", result.Decision)
	}
	if result.Winner != "A" {
		t.Errorf("Expected Winner=A, got %v", result.Winner)
	}
	if result.NA < result.NB {
		t.Errorf("Expected NA > NB since A beats B, got NA=%d, NB=%d", result.NA, result.NB)
	}
}

func TestSPRTEngineAcceptH0(t *testing.T) {
	engine := NewSPRTEngineWithDefaults()

	// B consistently beats A - clear evidence that A is not better
	scoresA := []float64{
		0.3, 0.35, 0.28, 0.32, 0.29,
		0.31, 0.33, 0.30, 0.27, 0.34,
		0.28, 0.36, 0.31, 0.29, 0.32,
		0.30, 0.33, 0.28, 0.31, 0.35,
	}
	scoresB := []float64{
		0.7, 0.65, 0.72, 0.68, 0.71,
		0.69, 0.67, 0.70, 0.73, 0.66,
		0.72, 0.64, 0.69, 0.71, 0.68,
		0.70, 0.67, 0.72, 0.69, 0.65,
	}

	result := engine.Test(scoresA, scoresB)
	// B beats A consistently, so we should accept null (B is better)
	if result.Decision != Accept {
		t.Errorf("Expected Accept when B is better, got %v", result.Decision)
	}
	if result.Winner != "B" {
		t.Errorf("Expected Winner=B, got %v", result.Winner)
	}
}

func TestSPRTEngineBBeatsA(t *testing.T) {
	engine := NewSPRTEngineWithDefaults()

	// B consistently beats A
	scoresA := []float64{
		0.3, 0.25, 0.28, 0.32, 0.35,
		0.29, 0.26, 0.31, 0.27, 0.33,
		0.28, 0.34, 0.30, 0.26, 0.29,
		0.31, 0.28, 0.32, 0.27, 0.30,
	}
	scoresB := []float64{
		0.9, 0.85, 0.88, 0.92, 0.87,
		0.91, 0.86, 0.89, 0.90, 0.88,
		0.87, 0.92, 0.85, 0.91, 0.86,
		0.89, 0.90, 0.88, 0.87, 0.91,
	}

	result := engine.Test(scoresA, scoresB)
	if result.Decision != Accept {
		t.Errorf("Expected Accept (B is better), got %v", result.Decision)
	}
	if result.Winner != "B" {
		t.Errorf("Expected Winner=B, got %v", result.Winner)
	}
}

func TestSPRTEngineMaxSamplesReached(t *testing.T) {
	config := SPRTConfig{
		Alpha:     0.05,
		Beta:      0.20,
		MinSamples: 15,
		MaxSamples: 30,
		P0:        0.5,
		P1:        0.6,
	}
	engine := NewSPRTEngine(config)

	// Create exactly 30 samples (at max)
	scoresA := make([]float64, 30)
	scoresB := make([]float64, 30)
	for i := 0; i < 30; i++ {
		scoresA[i] = 0.5 + float64(i)*0.01 // Slightly increasing
		scoresB[i] = 0.5 - float64(i)*0.01 // Slightly decreasing
	}

	result := engine.Test(scoresA, scoresB)
	// With 30 samples at max, should make a final decision (not Continue)
	if result.N != 30 {
		t.Errorf("Expected N=30, got %v", result.N)
	}
	// Should either Accept or Reject, not Continue
	if result.Decision == Continue {
		t.Errorf("Expected Accept or Reject at MaxSamples, got Continue")
	}
}

func TestSPRTEngineScoresCalculation(t *testing.T) {
	engine := NewSPRTEngineWithDefaults()

	scoresA := []float64{0.8, 0.7, 0.6}
	scoresB := []float64{0.5, 0.4, 0.3}

	result := engine.Test(scoresA, scoresB)

	expectedAvgA := (0.8 + 0.7 + 0.6) / 3.0
	expectedAvgB := (0.5 + 0.4 + 0.3) / 3.0

	if math.Abs(result.ScoreA-expectedAvgA) > 0.001 {
		t.Errorf("Expected ScoreA=%v, got %v", expectedAvgA, result.ScoreA)
	}
	if math.Abs(result.ScoreB-expectedAvgB) > 0.001 {
		t.Errorf("Expected ScoreB=%v, got %v", expectedAvgB, result.ScoreB)
	}
}

func TestSPRTEngineConfidenceInterval(t *testing.T) {
	engine := NewSPRTEngineWithDefaults()

	// A beats B 12 times out of 20
	scoresA := []float64{
		0.8, 0.9, 0.7, 0.85, 0.75,
		0.88, 0.92, 0.65, 0.78, 0.82,
		0.9, 0.8, 0.7, 0.85, 0.75,
		0.88, 0.92, 0.65, 0.78, 0.82,
	}
	scoresB := []float64{
		0.3, 0.4, 0.5, 0.35, 0.45,
		0.32, 0.38, 0.55, 0.42, 0.48,
		0.3, 0.4, 0.5, 0.35, 0.45,
		0.32, 0.38, 0.55, 0.42, 0.48,
	}

	result := engine.Test(scoresA, scoresB)

	// CI should be populated
	if result.ConfidenceCI[0] == 0 && result.ConfidenceCI[1] == 0 {
		t.Errorf("Expected non-zero ConfidenceCI")
	}
}

func TestSPRTEngineLogLikelihoodRatio(t *testing.T) {
	engine := NewSPRTEngineWithDefaults()

	// Test when A wins (isAWin=1)
	llr1 := engine.logLikelihoodRatio(1)
	// Expected: log(p1/p0) where p1=0.6, p0=0.5
	expectedLLR1 := math.Log(0.6 / 0.5)
	if math.Abs(llr1-expectedLLR1) > 0.001 {
		t.Errorf("Expected LLR(1)=%v, got %v", expectedLLR1, llr1)
	}

	// Test when A loses (isAWin=0)
	llr0 := engine.logLikelihoodRatio(0)
	// Expected: log((1-p1)/(1-p0)) where p1=0.6, p0=0.5
	expectedLLR0 := math.Log(0.4 / 0.5)
	if math.Abs(llr0-expectedLLR0) > 0.001 {
		t.Errorf("Expected LLR(0)=%v, got %v", expectedLLR0, llr0)
	}
}

func TestSPRTEnginePValue(t *testing.T) {
	engine := NewSPRTEngineWithDefaults()

	// Test with large log-likelihood ratio (strong evidence)
	pVal := engine.calculatePValue(5.0, 30)
	if pVal >= 0.05 {
		t.Errorf("Expected p-value < 0.05 for strong evidence, got %v", pVal)
	}

	// Test with small log-likelihood ratio (weak evidence)
	pVal2 := engine.calculatePValue(0.1, 30)
	if pVal2 < 0.05 {
		t.Errorf("Expected p-value >= 0.05 for weak evidence, got %v", pVal2)
	}
}

func TestSPRTEngineNAAndNB(t *testing.T) {
	engine := NewSPRTEngineWithDefaults()

	// A wins 3 times, B wins 2 times
	scoresA := []float64{0.8, 0.9, 0.7, 0.3, 0.4}
	scoresB := []float64{0.5, 0.4, 0.3, 0.6, 0.7}

	result := engine.Test(scoresA, scoresB)

	if result.N != 5 {
		t.Errorf("Expected N=5, got %v", result.N)
	}
	if result.NA != 3 {
		t.Errorf("Expected NA=3, got %v", result.NA)
	}
	if result.NB != 2 {
		t.Errorf("Expected NB=2, got %v", result.NB)
	}
}

func TestSPRTEngineMinMaxSamplesConfig(t *testing.T) {
	config := SPRTConfig{
		Alpha:     0.05,
		Beta:      0.20,
		MinSamples: 5,
		MaxSamples: 10,
		P0:        0.5,
		P1:        0.6,
	}
	engine := NewSPRTEngine(config)

	// With MinSamples=5 and MaxSamples=10, even 5 samples should trigger decision
	// Create a clear winner scenario
	scoresA := []float64{0.9, 0.85, 0.88, 0.92, 0.87}
	scoresB := []float64{0.3, 0.25, 0.28, 0.32, 0.35}

	result := engine.Test(scoresA, scoresB)

	// At exactly MinSamples, it should make a decision
	// (assuming strong evidence)
	if result.N != 5 {
		t.Errorf("Expected N=5, got %v", result.N)
	}
}

func TestSPRTEngineDifferentP0P1(t *testing.T) {
	config := SPRTConfig{
		Alpha:     0.05,
		Beta:      0.20,
		MinSamples: 10,
		MaxSamples: 20,
		P0:        0.3, // Lower baseline
		P1:        0.5, // Higher target
	}
	engine := NewSPRTEngine(config)

	// With P0=0.3, P1=0.5, A should win more often (need 10+ samples for decision)
	scoresA := []float64{
		0.8, 0.75, 0.82, 0.78, 0.85,
		0.79, 0.81, 0.76, 0.83, 0.77,
		0.80, 0.74, 0.84, 0.79, 0.82,
	}
	scoresB := []float64{
		0.2, 0.15, 0.18, 0.22, 0.12,
		0.19, 0.21, 0.17, 0.23, 0.13,
		0.20, 0.16, 0.14, 0.18, 0.15,
	}

	result := engine.Test(scoresA, scoresB)

	// A should be the winner due to strong evidence
	if result.Decision != Reject {
		t.Errorf("Expected Reject (A wins), got %v", result.Decision)
	}
	if result.Winner != "A" {
		t.Errorf("Expected Winner=A, got %v", result.Winner)
	}
}
