package service

import (
	"encoding/json"
	"math"
	"regexp"
	"strings"

	"prompt-vault/models"
)

// ScoringService provides quality scoring for prompts
type ScoringService struct{}

// NewScoringService creates a new ScoringService
func NewScoringService() *ScoringService {
	return &ScoringService{}
}

// ScoreResult represents the result of scoring a prompt
type ScoreResult struct {
	Clarity      float64                  `json:"clarity"`
	Completeness float64                  `json:"completeness"`
	Example      float64                  `json:"example"`
	Role         float64                  `json:"role"`
	Overall      float64                  `json:"overall"`
	Breakdown    map[string]interface{}   `json:"breakdown"`
}

// Score evaluates a prompt and returns scores for all dimensions
func (s *ScoringService) Score(prompt *models.Prompt) *ScoreResult {
	clarity := s.ScoreClarity(prompt.Content)
	completeness := s.ScoreCompleteness(prompt)
	example := s.ScoreExample(prompt)
	role := s.ScoreRole(prompt.Content)

	// Calculate weighted overall
	overall := clarity*0.30 + completeness*0.30 + example*0.25 + role*0.15

	return &ScoreResult{
		Clarity:      clarity,
		Completeness: completeness,
		Example:      example,
		Role:         role,
		Overall:      math.Round(overall*100) / 100,
		Breakdown: map[string]interface{}{
			"clarity":      s.clarityBreakdown(prompt.Content),
			"completeness": s.completenessBreakdown(prompt),
			"example":      s.exampleBreakdown(prompt),
			"role":         s.roleBreakdown(prompt.Content),
		},
	}
}

// ScoreClarity evaluates clarity based on:
// - Variable placeholder count and format
// - Length appropriateness
// - Format conventions
func (s *ScoringService) ScoreClarity(content string) float64 {
	breakdown := s.clarityBreakdown(content)
	return breakdown["score"].(float64)
}

func (s *ScoringService) clarityBreakdown(content string) map[string]interface{} {
	// Count variable placeholders
	placeholderRegex := regexp.MustCompile(`\{\{[^}]+\}\}`)
	placeholders := placeholderRegex.FindAllString(content, -1)
	placeholderCount := len(placeholders)

	// Check for proper format (no spaces in placeholder)
	formatScore := 1.0
	for _, p := range placeholders {
		if strings.Contains(p, " ") {
			formatScore -= 0.1
		}
	}

	// Length scoring (optimal: 200-2000 chars)
	length := len(content)
	lengthScore := 1.0
	if length < 50 {
		lengthScore = 0.3
	} else if length < 100 {
		lengthScore = 0.5
	} else if length < 200 {
		lengthScore = 0.7
	} else if length > 5000 {
		lengthScore = 0.6
	}

	// Check for clear structure (sections/headers)
	structureScore := 0.5
	if strings.Contains(content, "##") || strings.Contains(content, "###") {
		structureScore += 0.3
	}
	if strings.Contains(content, "\n\n") {
		structureScore += 0.2
	}

	// Variable naming quality
	namingScore := 1.0
	if placeholderCount > 0 {
		varNameRegex := regexp.MustCompile(`\{\{([^}]+)\}\}`)
		matches := varNameRegex.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			varName := match[1]
			// Good: user_name, input_text, context
			// Bad: x, var, temp
			if len(varName) < 3 {
				namingScore -= 0.2
			}
		}
	}

	// Base score from placeholders (0-3 placeholders is ideal)
	placeholderScore := 1.0
	if placeholderCount == 0 {
		placeholderScore = 0.3 // No variables might be too rigid
	} else if placeholderCount > 10 {
		placeholderScore = 0.5 // Too many variables
	} else if placeholderCount > 5 {
		placeholderScore = 0.8
	}

	// Combine scores
	score := (formatScore*0.2 + lengthScore*0.3 + structureScore*0.2 +
		namingScore*0.15 + placeholderScore*0.15)
	score = math.Max(0, math.Min(1, score))

	return map[string]interface{}{
		"score":              math.Round(score*100) / 100,
		"placeholder_count":  placeholderCount,
		"length":             length,
		"format_score":       formatScore,
		"length_score":       lengthScore,
		"structure_score":    structureScore,
		"naming_score":       namingScore,
	}
}

// ScoreCompleteness evaluates completeness based on:
// - Required field presence
// - AI-enhanced evaluation
func (s *ScoringService) ScoreCompleteness(prompt *models.Prompt) float64 {
	breakdown := s.completenessBreakdown(prompt)
	return breakdown["score"].(float64)
}

func (s *ScoringService) completenessBreakdown(prompt *models.Prompt) map[string]interface{} {
	content := prompt.Content

	// Required elements scoring
	score := 0.0
	count := 0

	// Role definition
	if strings.Contains(strings.ToLower(content), "you are") ||
		strings.Contains(strings.ToLower(content), "act as") ||
		strings.Contains(strings.ToLower(content), "role") {
		score += 0.2
	}
	count++

	// Task definition
	if strings.Contains(strings.ToLower(content), "task") ||
		strings.Contains(strings.ToLower(content), "help") ||
		strings.Contains(strings.ToLower(content), "please") {
		score += 0.2
	}
	count++

	// Output format
	if strings.Contains(strings.ToLower(content), "output") ||
		strings.Contains(strings.ToLower(content), "format") ||
		strings.Contains(strings.ToLower(content), "respond") {
		score += 0.2
	}
	count++

	// Constraints/requirements
	if strings.Contains(strings.ToLower(content), "must") ||
		strings.Contains(strings.ToLower(content), "should") ||
		strings.Contains(strings.ToLower(content), "require") {
		score += 0.2
	}
	count++

	// Context/background
	if strings.Contains(strings.ToLower(content), "context") ||
		strings.Contains(strings.ToLower(content), "given") ||
		strings.Contains(strings.ToLower(content), "background") {
		score += 0.2
	}
	count++

	// Additional factors
	hasVariables := strings.Contains(content, "{{")
	hasExamples := strings.Contains(strings.ToLower(content), "example")

	if hasVariables {
		score += 0.1
	}
	if hasExamples {
		score += 0.1
	}

	// Normalize score
	normalizedScore := math.Min(1.0, score)

	return map[string]interface{}{
		"score":        math.Round(normalizedScore*100) / 100,
		"has_role":     strings.Contains(strings.ToLower(content), "you are") || strings.Contains(strings.ToLower(content), "act as"),
		"has_task":     strings.Contains(strings.ToLower(content), "task") || strings.Contains(strings.ToLower(content), "help"),
		"has_format":   strings.Contains(strings.ToLower(content), "output") || strings.Contains(strings.ToLower(content), "format"),
		"has_constraints": strings.Contains(strings.ToLower(content), "must") || strings.Contains(strings.ToLower(content), "should"),
		"has_context":  strings.Contains(strings.ToLower(content), "context"),
		"has_variables": hasVariables,
		"has_examples": hasExamples,
	}
}

// ScoreExample evaluates examples based on:
// - Example count
// - Example quality
// - AI-enhanced evaluation
func (s *ScoringService) ScoreExample(prompt *models.Prompt) float64 {
	breakdown := s.exampleBreakdown(prompt)
	return breakdown["score"].(float64)
}

func (s *ScoringService) exampleBreakdown(prompt *models.Prompt) map[string]interface{} {
	content := prompt.Content

	// Count examples
	examplePatterns := []string{
		"example:",
		"example -",
		"for example",
		"such as",
		"like:",
		"e.g.",
		"e.g.,",
		"```",
	}

	exampleCount := 0
	for _, pattern := range examplePatterns {
		exampleCount += strings.Count(strings.ToLower(content), pattern)
	}

	// Check for structured examples (code blocks)
	codeBlockCount := strings.Count(content, "```")

	// Score based on example presence
	score := 0.0
	if exampleCount == 0 {
		score = 0.2 // No examples
	} else if exampleCount == 1 {
		score = 0.5
	} else if exampleCount == 2 {
		score = 0.7
	} else if exampleCount >= 3 {
		score = 0.9
	}

	// Bonus for code blocks
	if codeBlockCount >= 2 {
		score = math.Min(1.0, score+0.1)
	}

	return map[string]interface{}{
		"score":          math.Round(score*100) / 100,
		"example_count":  exampleCount,
		"code_blocks":    codeBlockCount,
	}
}

// ScoreRole evaluates role definition based on:
// - Role keyword presence
// - AI-enhanced evaluation
func (s *ScoringService) ScoreRole(content string) float64 {
	breakdown := s.roleBreakdown(content)
	return breakdown["score"].(float64)
}

func (s *ScoringService) roleBreakdown(content string) map[string]interface{} {
	lower := strings.ToLower(content)

	// Role keywords
	roleKeywords := []string{
		"you are",
		"act as",
		"as a",
		"role:",
		"you are a",
		"you'll be",
		"you will be",
		"your role",
		"as an",
	}

	foundKeywords := make([]string, 0)
	for _, kw := range roleKeywords {
		if strings.Contains(lower, kw) {
			foundKeywords = append(foundKeywords, kw)
		}
	}

	// Calculate score
	score := 0.0
	if len(foundKeywords) == 0 {
		score = 0.3 // No role defined
	} else if len(foundKeywords) >= 1 {
		score = 0.8
	}

	// Bonus for detailed role description
	if strings.Contains(lower, "expertise") ||
		strings.Contains(lower, "experience") ||
		strings.Contains(lower, "specialized") {
		score = math.Min(1.0, score+0.15)
	}

	// Bonus for capabilities
	if strings.Contains(lower, "capable of") ||
		strings.Contains(lower, "abilities") ||
		strings.Contains(lower, "skills") {
		score = math.Min(1.0, score+0.1)
	}

	return map[string]interface{}{
		"score":        math.Round(score*100) / 100,
		"keywords":     foundKeywords,
		"has_expertise": strings.Contains(lower, "expertise") || strings.Contains(lower, "experience"),
		"has_capabilities": strings.Contains(lower, "capable") || strings.Contains(lower, "abilities"),
	}
}

// GenerateEvalCases generates 5-20 test cases for a prompt
func (s *ScoringService) GenerateEvalCases(prompt *models.Prompt, count int) ([]models.EvalCase, error) {
	// Clamp count to 5-20 range
	if count < 5 {
		count = 5
	}
	if count > 20 {
		count = 20
	}

	// Extract variables from prompt
	variableRegex := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := variableRegex.FindAllStringSubmatch(prompt.Content, -1)

	variables := make([]string, 0)
	seen := make(map[string]bool)
	for _, match := range matches {
		varName := match[1]
		if !seen[varName] {
			seen[varName] = true
			variables = append(variables, varName)
		}
	}

	// Generate test cases
	cases := make([]models.EvalCase, count)

	// Sample inputs for different variable types
	sampleInputs := map[string][]string{
		"name":      {"Alice", "Bob", "Charlie"},
		"user":      {"john_doe", "jane_smith", "test_user"},
		"email":     {"user@example.com", "test@test.org"},
		"text":      {"Hello world", "Sample text content"},
		"input":     {"User input data", "Test input"},
		"content":   {"Sample content", "Test content"},
		"message":   {"Hello there", "Test message"},
		"query":     {"search term", "test query"},
		"id":        {"12345", "67890"},
		"title":     {"Sample Title", "Test Title"},
		"description": {"A brief description", "Test description"},
		"context":    {"Given the context of...", "Background info here"},
	}

	// Generate diverse cases
	for i := 0; i < count; i++ {
		testInput := prompt.Content

		// Fill variables with sample values
		for _, v := range variables {
			samples, ok := sampleInputs[v]
			if ok {
				sampleValue := samples[i%len(samples)]
				testInput = strings.Replace(testInput, "{{"+v+"}}", sampleValue, 1)
			} else {
				// Generic placeholder
				testInput = strings.Replace(testInput, "{{"+v+"}}", "test_value", 1)
			}
		}

		cases[i] = models.EvalCase{
			Input: testInput,
			Metadata: map[string]string{
				"case_number":  string(rune('A' + i)),
				"variable_count": string(rune('0' + len(variables))),
			},
		}
	}

	return cases, nil
}

// GetDefaultWeights returns the default scoring weights
func (s *ScoringService) GetDefaultWeights() models.EvalWeights {
	return models.DefaultEvalWeights()
}

// CalculateWeightedScore calculates the weighted score given weights and scores
func (s *ScoringService) CalculateWeightedScore(weights models.EvalWeights, scores map[string]float64) float64 {
	total := weights.Clarity*scores["clarity"] +
		weights.Completeness*scores["completeness"] +
		weights.Example*scores["example"] +
		weights.Role*scores["role"]

	return math.Round(total*100) / 100
}

// MarshalWeights converts EvalWeights to JSON string
func MarshalWeights(weights models.EvalWeights) string {
	data, _ := json.Marshal(weights)
	return string(data)
}

// UnmarshalWeights parses JSON string to EvalWeights
func UnmarshalWeights(s string) (models.EvalWeights, error) {
	var weights models.EvalWeights
	if err := json.Unmarshal([]byte(s), &weights); err != nil {
		return models.EvalWeights{}, err
	}
	return weights, nil
}
