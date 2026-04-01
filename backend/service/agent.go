package service

import (
	"fmt"

	"prompt-vault/models"

	"gorm.io/gorm"
)

// AgentService encapsulates business logic for Agent entities.
type AgentService struct {
	db *gorm.DB
}

// NewAgentService creates a new AgentService.
func NewAgentService(db *gorm.DB) *AgentService {
	return &AgentService{db: db}
}

// Clone creates a copy of an agent with "(Copy)" appended to the name.
// The cloned agent always has source set to "custom".
func (s *AgentService) Clone(agentID uint) (*models.Agent, error) {
	var agent models.Agent
	if err := s.db.First(&agent, agentID).Error; err != nil {
		return nil, err
	}

	clone := models.Agent{
		Name:         agent.Name + " (Copy)",
		Role:         agent.Role,
		Content:      agent.Content,
		ContentCN:    agent.ContentCN,
		Capabilities: agent.Capabilities,
		Category:     agent.Category,
		Source:       "custom",
	}

	if err := s.db.Create(&clone).Error; err != nil {
		return nil, err
	}
	return &clone, nil
}

// Delete removes an agent by ID. Returns gorm.ErrRecordNotFound if not found.
func (s *AgentService) Delete(agentID uint) error {
	result := s.db.Delete(&models.Agent{}, agentID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Count returns the total number of agents.
func (s *AgentService) Count() (int64, error) {
	var count int64
	err := s.db.Model(&models.Agent{}).Count(&count).Error
	return count, err
}

// GetByID retrieves an agent by ID. Returns gorm.ErrRecordNotFound if not found.
func (s *AgentService) GetByID(agentID uint) (*models.Agent, error) {
	var agent models.Agent
	if err := s.db.First(&agent, agentID).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

// CloneWithActivity wraps Clone and returns a formatted details JSON string.
func (s *AgentService) CloneWithActivity(agentID uint) (*models.Agent, string, error) {
	clone, err := s.Clone(agentID)
	if err != nil {
		return nil, "", err
	}
	return clone, fmt.Sprintf(`{"from_id": %d}`, agentID), nil
}
