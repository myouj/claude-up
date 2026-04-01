package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/middleware"
	"prompt-vault/models"
)

type SettingHandler struct {
	db *gorm.DB
}

func NewSettingHandler(db *gorm.DB) *SettingHandler {
	return &SettingHandler{db: db}
}

func (h *SettingHandler) Get(c *gin.Context) {
	key := c.Param("key")

	var setting models.Setting
	if err := h.db.First(&setting, "key = ?", key).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Setting not found"})
		return
	}

	value := setting.Value
	if setting.IsSecret {
		value = decryptValue(value)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"key":       setting.Key,
			"value":     value,
			"is_secret": setting.IsSecret,
		},
	})
}

func (h *SettingHandler) Set(c *gin.Context) {
	key := c.Param("key")

	var input struct {
		Value    string `json:"value" binding:"required"`
		IsSecret bool   `json:"is_secret"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	value := input.Value
	if input.IsSecret {
		value = encryptValue(value)
	}

	existing := models.Setting{}
	if err := h.db.First(&existing, "key = ?", key).Error; err == nil {
		h.db.Model(&existing).Updates(map[string]interface{}{
			"value":     value,
			"is_secret": input.IsSecret,
		})
	} else {
		h.db.Create(&models.Setting{
			Key:      key,
			Value:    value,
			IsSecret: input.IsSecret,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Setting saved",
	})
}

func (h *SettingHandler) Delete(c *gin.Context) {
	key := c.Param("key")

	if err := h.db.Delete(&models.Setting{}, "key = ?", key).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Setting deleted",
	})
}

func (h *SettingHandler) List(c *gin.Context) {
	var settings []models.Setting
	h.db.Order("key ASC").Find(&settings)

	var responses []map[string]interface{}
	for _, s := range settings {
		value := s.Value
		if s.IsSecret && value != "" {
			value = "********"
		}
		responses = append(responses, map[string]interface{}{
			"key":       s.Key,
			"value":     value,
			"is_secret": s.IsSecret,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
	})
}

// GetAllAPIKeys returns all API keys (decrypted) for provider use.
func (h *SettingHandler) GetAllAPIKeys() map[string]string {
	return map[string]string{
		"openai":  getEnvOrSetting(h.db, "openai_api_key"),
		"claude":  getEnvOrSetting(h.db, "anthropic_api_key"),
		"gemini":  getEnvOrSetting(h.db, "gemini_api_key"),
		"minimax": getEnvOrSetting(h.db, "minimax_api_key"),
	}
}

func (h *SettingHandler) GetByKey(key string) (*models.Setting, error) {
	var setting models.Setting
	if err := h.db.Where("key = ?", key).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func getEnvOrSetting(db *gorm.DB, key string) string {
	if val := os.Getenv(strings.ToUpper(key)); val != "" {
		return val
	}
	var setting models.Setting
	if err := db.First(&setting, "key = ?", key).Error; err == nil {
		return decryptValue(setting.Value)
	}
	return ""
}

// ----- AES-256-GCM Encryption -----

func getEncryptionKey() []byte {
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		middleware.Fatal("ENCRYPTION_KEY environment variable is not set", map[string]interface{}{
			"hint": "Set ENCRYPTION_KEY to a 32-byte (256-bit) secret before starting the server",
		})
	}
	b := []byte(key)
	if len(b) < 32 {
		middleware.Fatal("ENCRYPTION_KEY is too short", map[string]interface{}{
			"hint":    "ENCRYPTION_KEY must be at least 32 bytes (256 bits) long",
			"current": len(b),
		})
	}
	if len(b) > 32 {
		return b[:32]
	}
	return b
}

func encryptValue(plaintext string) string {
	if plaintext == "" {
		return ""
	}
	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return plaintext
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return plaintext
	}

	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)
	return string(gcm.Seal(nonce, nonce, []byte(plaintext), nil))
}

func decryptValue(ciphertext string) string {
	if ciphertext == "" {
		return ""
	}
	data := []byte(ciphertext)
	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return ciphertext
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return ciphertext
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return ciphertext
	}

	nonce, ct := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return ciphertext // failed to decrypt, return as-is
	}
	return string(plaintext)
}
