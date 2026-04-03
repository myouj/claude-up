package service

import (
	"math"
)

// SPRTConfig holds configuration for the SPRT test.
type SPRTConfig struct {
	Alpha     float64 // Significance level (Type I error), default 0.05
	Beta      float64 // Power level (1 - Type II error), default 0.20
	MinSamples int    // Minimum samples before making decision, default 15
	MaxSamples int    // Maximum samples, default 50
	P0        float64 // Null hypothesis success rate, default 0.5
	P1        float64 // Alternative hypothesis success rate, default 0.6
}

// Decision represents the SPRT test decision.
type Decision int

const (
	Continue Decision = iota // Continue sampling
	Accept                   // Accept null hypothesis (H0)
	Reject                   // Reject null hypothesis (H1 is better)
)

// SPRTResult holds the result of an SPRT test.
type SPRTResult struct {
	Decision     Decision   `json:"decision"`      // Continue, Accept, or Reject
	N            int        `json:"n"`             // Total samples
	NA           int        `json:"n_a"`           // Samples where A wins
	NB           int        `json:"n_b"`           // Samples where B wins
	ScoreA       float64    `json:"score_a"`       // Average score for A
	ScoreB       float64    `json:"score_b"`       // Average score for B
	PValue       float64    `json:"p_value"`       // Approximate p-value
	ConfidenceCI [2]float64 `json:"ci"`            // 95% confidence interval for difference
	Winner       string     `json:"winner"`        // "A", "B", or ""
}

// SPRTEngine performs Sequential Probability Ratio Tests for A/B testing.
type SPRTEngine struct {
	config SPRTConfig
}

// NewSPRTEngine creates a new SPRT engine with the given configuration.
func NewSPRTEngine(config SPRTConfig) *SPRTEngine {
	// Set defaults
	if config.Alpha == 0 {
		config.Alpha = 0.05
	}
	if config.Beta == 0 {
		config.Beta = 0.20
	}
	if config.MinSamples == 0 {
		config.MinSamples = 15
	}
	if config.MaxSamples == 0 {
		config.MaxSamples = 50
	}
	if config.P0 == 0 {
		config.P0 = 0.5
	}
	if config.P1 == 0 {
		config.P1 = 0.6
	}

	return &SPRTEngine{config: config}
}

// NewSPRTEngineWithDefaults creates a new SPRT engine with default configuration.
func NewSPRTEngineWithDefaults() *SPRTEngine {
	return NewSPRTEngine(SPRTConfig{
		Alpha:     0.05,
		Beta:      0.20,
		MinSamples: 15,
		MaxSamples: 50,
		P0:        0.5,
		P1:        0.6,
	})
}

// Test performs SPRT analysis on the given scores.
// scoresA and scoresB are parallel arrays of scores for variants A and B.
// Each position represents one test/sample comparing A vs B.
func (e *SPRTEngine) Test(scoresA, scoresB []float64) SPRTResult {
	n := len(scoresA)
	if n == 0 || len(scoresB) == 0 {
		return SPRTResult{
			Decision: Continue,
			N:        0,
			Winner:   "",
		}
	}

	// Count wins for A (scoreA > scoreB)
	nA := 0
	var totalA, totalB float64
	for i := 0; i < n && i < len(scoresB); i++ {
		totalA += scoresA[i]
		totalB += scoresB[i]
		if scoresA[i] > scoresB[i] {
			nA++
		}
	}

	// Clamp n to the actual number of paired comparisons
	if len(scoresB) < n {
		n = len(scoresB)
	}
	if n == 0 {
		return SPRTResult{
			Decision: Continue,
			N:        0,
			Winner:   "",
		}
	}

	// Calculate average scores
	avgA := totalA / float64(n)
	avgB := totalB / float64(n)

	// Calculate boundaries
	logA := math.Log(e.config.Beta / (1 - e.config.Alpha))
	logB := math.Log((1 - e.config.Beta) / e.config.Alpha)

	// Calculate cumulative log-likelihood ratio using individual observations
	var sumLogLR float64
	for i := 0; i < n; i++ {
		if i >= len(scoresA) || i >= len(scoresB) {
			break
		}
		// Success (1) if A beats B at this observation
		isAWin := 0
		if scoresA[i] > scoresB[i] {
			isAWin = 1
		}
		sumLogLR += e.logLikelihoodRatio(isAWin)
	}

	// Make decision based on boundaries
	decision := Continue
	if n >= e.config.MinSamples {
		if sumLogLR <= logA {
			decision = Accept // Accept null: A is not better than B
		} else if sumLogLR >= logB {
			decision = Reject // Reject null: A is better than B
		}
	}

	// If we reached max samples, make final decision
	if n >= e.config.MaxSamples && decision == Continue {
		// Use the final likelihood ratio to decide
		if sumLogLR >= 0 {
			decision = Reject // A is better
		} else {
			decision = Accept // B is better or equal
		}
	}

	// Calculate p-value (approximate, using asymptotic distribution)
	pValue := e.calculatePValue(sumLogLR, n)

	// Calculate confidence interval for the difference
	ci := e.calculateCI(nA, n, avgA-avgB)

	// Determine winner
	winner := ""
	if decision == Reject {
		winner = "A"
	} else if decision == Accept {
		winner = "B"
	}

	return SPRTResult{
		Decision:     decision,
		N:            n,
		NA:           nA,
		NB:           n - nA,
		ScoreA:       avgA,
		ScoreB:       avgB,
		PValue:       pValue,
		ConfidenceCI: ci,
		Winner:       winner,
	}
}

// logLikelihoodRatio calculates the log-likelihood ratio for a single observation.
// isAWin is 1 if variant A wins at this observation, 0 otherwise.
func (e *SPRTEngine) logLikelihoodRatio(isAWin int) float64 {
	p0 := e.config.P0
	p1 := e.config.P1

	// Likelihood under H1 (alternative: success rate is p1)
	likelihood1 := math.Pow(p1, float64(isAWin)) * math.Pow(1-p1, 1-float64(isAWin))
	// Likelihood under H0 (null: success rate is p0)
	likelihood0 := math.Pow(p0, float64(isAWin)) * math.Pow(1-p0, 1-float64(isAWin))

	if likelihood0 <= 0 || likelihood1 <= 0 {
		return 0
	}

	return math.Log(likelihood1 / likelihood0)
}

// calculatePValue approximates the p-value from the log-likelihood ratio.
// This is a simplified approximation suitable for sequential testing.
func (e *SPRTEngine) calculatePValue(logLR float64, n int) float64 {
	// Under null hypothesis, 2*logLR approximately follows chi-squared with 1 df
	// This is the sequential generalization of the likelihood ratio test
	chi2 := 2 * math.Abs(logLR)
	if chi2 < 0 {
		chi2 = 0
	}

	// Approximate p-value using chi-squared distribution (1 df)
	// For large n, this is a reasonable approximation
	pValue := math.Exp(-chi2 / 2)
	if pValue > 1 {
		pValue = 1
	}

	return pValue
}

// calculateCI calculates the 95% confidence interval for the difference in proportions.
// Returns [lower, upper] bounds.
func (e *SPRTEngine) calculateCI(nA, n int, diff float64) [2]float64 {
	if n == 0 {
		return [2]float64{0, 0}
	}

	// Observed proportion
	pHat := float64(nA) / float64(n)

	// Standard error of proportion
	se := math.Sqrt(pHat*(1-pHat) / float64(n))
	if se < 1e-10 {
		se = 1e-10
	}

	// 95% CI using normal approximation
	z := 1.96
	lower := diff - z*se
	upper := diff + z*se

	return [2]float64{lower, upper}
}
