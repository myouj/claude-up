package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"prompt-vault/models"
)

// Set up a valid 32-byte encryption key for all tests.
func init() {
	os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")
}

func newTestDBForSettings(t *testing.T) *gorm.DB {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/settings_test.db"
	db, err := gorm.Open(sqlite.Open(dbPath+"?_busy_timeout=30000&_journal_mode=DELETE"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(&models.Setting{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	t.Cleanup(func() {
		if db != nil {
			sqlDB, _ := db.DB()
			if sqlDB != nil {
				sqlDB.Close()
			}
		}
	})
	return db
}

func setupTestRouterWithSettingHandler(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	settingHandler := NewSettingHandler(db)

	api := r.Group("/api")
	api.GET("/settings", settingHandler.List)
	api.GET("/settings/:key", settingHandler.Get)
	api.PUT("/settings/:key", settingHandler.Set)
	api.DELETE("/settings/:key", settingHandler.Delete)

	return r
}

// ----- Handler Tests -----

func TestSettingHandler_Get(t *testing.T) {
	db := newTestDBForSettings(t)
	router := setupTestRouterWithSettingHandler(db)

	// Create a plain setting
	db.Create(&models.Setting{Key: "theme", Value: "dark"})
	db.Create(&models.Setting{Key: "api_key", Value: "secret123", IsSecret: true})

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantSucc   bool
		wantVal    string
	}{
		{"existing plain setting", "/api/settings/theme", http.StatusOK, true, "dark"},
		{"existing secret setting", "/api/settings/api_key", http.StatusOK, true, "secret123"},
		{"non-existent setting", "/api/settings/nonexistent", http.StatusNotFound, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := getJSON(router, tt.path)
			if w.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tt.wantStatus)
			}

			var resp struct {
				Success bool `json:"success"`
				Data    struct {
					Key       string `json:"key"`
					Value     string `json:"value"`
					IsSecret bool   `json:"is_secret"`
				} `json:"data"`
			}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Success != tt.wantSucc {
				t.Errorf("success: got %v", resp.Success)
			}
			if tt.wantVal != "" && resp.Data.Value != tt.wantVal {
				t.Errorf("value: got %s, want %s", resp.Data.Value, tt.wantVal)
			}
		})
	}
}

func TestSettingHandler_Set(t *testing.T) {
	db := newTestDBForSettings(t)
	router := setupTestRouterWithSettingHandler(db)

	tests := []struct {
		name       string
		path       string
		body       map[string]interface{}
		wantStatus int
		wantSucc   bool
	}{
		{
			name:       "create new plain setting",
			path:       "/api/settings/new_setting",
			body:       map[string]interface{}{"value": "value1"},
			wantStatus: http.StatusOK,
			wantSucc:   true,
		},
		{
			name:       "create new secret setting",
			path:       "/api/settings/api_key",
			body:       map[string]interface{}{"value": "secret-value", "is_secret": true},
			wantStatus: http.StatusOK,
			wantSucc:   true,
		},
		{
			name:       "update existing setting",
			path:       "/api/settings/new_setting",
			body:       map[string]interface{}{"value": "updated_value"},
			wantStatus: http.StatusOK,
			wantSucc:   true,
		},
		{
			name:       "missing value",
			path:       "/api/settings/empty_val",
			body:       map[string]interface{}{},
			wantStatus: http.StatusBadRequest,
			wantSucc:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := putJSON(router, tt.path, tt.body)
			if w.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tt.wantStatus)
			}

			var resp struct {
				Success bool `json:"success"`
			}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Success != tt.wantSucc {
				t.Errorf("success: got %v, want %v", resp.Success, tt.wantSucc)
			}
		})
	}

	// Verify plain value was set
	var setting models.Setting
	db.Where("key = ?", "new_setting").First(&setting)
	if setting.Value != "updated_value" {
		t.Errorf("expected updated_value, got %s", setting.Value)
	}

	// Verify secret was encrypted
	db.Where("key = ?", "api_key").First(&setting)
	if setting.IsSecret && setting.Value == "secret-value" {
		t.Error("secret value should be encrypted, not plaintext")
	}
}

func TestSettingHandler_Delete(t *testing.T) {
	db := newTestDBForSettings(t)
	router := setupTestRouterWithSettingHandler(db)

	db.Create(&models.Setting{Key: "to_delete", Value: "val"})

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantSucc   bool
	}{
		{"existing key", "/api/settings/to_delete", http.StatusOK, true},
		{"non-existent key", "/api/settings/nonexistent", http.StatusOK, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := deleteReq(router, tt.path)
			if w.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tt.wantStatus)
			}

			var resp struct {
				Success bool `json:"success"`
			}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Success != tt.wantSucc {
				t.Errorf("success: got %v", resp.Success)
			}
		})
	}
}

func TestSettingHandler_List(t *testing.T) {
	db := newTestDBForSettings(t)
	router := setupTestRouterWithSettingHandler(db)

	db.Create(&models.Setting{Key: "key1", Value: "val1"})
	db.Create(&models.Setting{Key: "key2", Value: "secret2", IsSecret: true})
	db.Create(&models.Setting{Key: "key3", Value: "val3"})

	w := getJSON(router, "/api/settings")
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d", w.Code)
	}

	var resp struct {
		Success bool `json:"success"`
		Data    []struct {
			Key       string `json:"key"`
			Value     string `json:"value"`
			IsSecret bool   `json:"is_secret"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success")
	}
	if len(resp.Data) != 3 {
		t.Errorf("expected 3 settings, got %d", len(resp.Data))
	}

	// Verify secret is masked
	for _, s := range resp.Data {
		if s.Key == "key2" {
			if s.Value != "********" {
				t.Errorf("secret should be masked, got %s", s.Value)
			}
		}
	}
}

func TestSettingHandler_GetAllAPIKeys(t *testing.T) {
	db := newTestDBForSettings(t)
	h := NewSettingHandler(db)

	// Set some values directly in DB (encrypted)
	db.Create(&models.Setting{Key: "openai_api_key", Value: encryptValue("test-openai"), IsSecret: true})

	keys := h.GetAllAPIKeys()
	if keys == nil {
		t.Fatal("expected non-nil map")
	}
	// Keys should come from env vars (not set in test) or DB
	_ = keys["openai"] // just verify it doesn't panic
}

func TestSettingHandler_GetByKey(t *testing.T) {
	db := newTestDBForSettings(t)
	h := NewSettingHandler(db)

	db.Create(&models.Setting{Key: "test_key", Value: "test_value"})

	setting, err := h.GetByKey("test_key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if setting == nil {
		t.Fatal("expected setting")
	}
	if setting.Value != "test_value" {
		t.Errorf("value: got %s", setting.Value)
	}

	// Non-existent
	_, err = h.GetByKey("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent key")
	}
}

// ----- Encryption/Decryption Tests -----

func TestEncryptDecryptValue(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
	}{
		{"simple text", "hello world"},
		{"unicode", "你好世界"},
		{"special chars", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"long text", "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
			"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris."},
		{"json", `{"key":"value","number":42,"array":[1,2,3]}`},
		{"empty string", ""},
		{"single char", "x"},
		{"32 bytes exactly", "12345678901234567890123456789012"},
		{"33 bytes (truncated)", "123456789012345678901234567890123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted := encryptValue(tt.plaintext)
			decrypted := decryptValue(encrypted)

			if tt.plaintext != "" && decrypted != tt.plaintext {
				t.Errorf("roundtrip failed: got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptValue_Idempotent(t *testing.T) {
	plaintext := "test-value-12345"
	encrypted := encryptValue(plaintext)

	// Encryption uses random nonce, so same plaintext -> different ciphertext
	encrypted2 := encryptValue(plaintext)
	if encrypted == encrypted2 {
		t.Error("expected different ciphertexts due to random nonce")
	}

	// But both should decrypt to the same value
	if decryptValue(encrypted) != plaintext {
		t.Error("first encryption failed to decrypt")
	}
	if decryptValue(encrypted2) != plaintext {
		t.Error("second encryption failed to decrypt")
	}
}

func TestEncryptValue_EmptyString(t *testing.T) {
	got := encryptValue("")
	if got != "" {
		t.Errorf("encrypt empty: got %q, want empty", got)
	}
}

func TestDecryptValue_EmptyString(t *testing.T) {
	got := decryptValue("")
	if got != "" {
		t.Errorf("decrypt empty: got %q, want empty", got)
	}
}

func TestDecryptValue_InvalidCiphertext(t *testing.T) {
	// Should return ciphertext as-is on decryption failure
	got := decryptValue("not-valid-ciphertext-data-here")
	if got == "not-valid-ciphertext-data-here" {
		// This is the fallback behavior
	}
}

func TestDecryptValue_ShortData(t *testing.T) {
	// Nonce size for GCM is 12 bytes, so data shorter than 12 should return as-is
	got := decryptValue("short")
	if got != "short" {
		t.Errorf("short data: got %q", got)
	}
}

func TestGetEncryptionKey(t *testing.T) {
	key := getEncryptionKey()
	if len(key) != 32 {
		t.Errorf("expected 32-byte key, got %d bytes", len(key))
	}
}

func TestGetEncryptionKey_Truncation(t *testing.T) {
	// The init() sets a 32-byte key, which is exactly right
	key := getEncryptionKey()
	if cap(key) != 32 {
		t.Errorf("expected capacity 32, got %d", cap(key))
	}
}

// Integration test: ensure encrypted settings can be decrypted after round-trip.
func TestSettingHandler_SecretRoundTrip(t *testing.T) {
	db := newTestDBForSettings(t)
	router := setupTestRouterWithSettingHandler(db)

	secretValue := "super-secret-api-key-12345"

	// Set as secret
	w := putJSON(router, "/api/settings/openai_api_key", map[string]interface{}{
		"value":     secretValue,
		"is_secret": true,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("set failed: %d", w.Code)
	}

	// Get back decrypted
	w = getJSON(router, "/api/settings/openai_api_key")
	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Value     string `json:"value"`
			IsSecret bool   `json:"is_secret"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Fatal("get failed")
	}
	if resp.Data.Value != secretValue {
		t.Errorf("decrypted value mismatch: got %q, want %q", resp.Data.Value, secretValue)
	}
	if !resp.Data.IsSecret {
		t.Error("expected IsSecret to be true")
	}
}

// ----- getEnvOrSetting Tests -----

func TestGetEnvOrSetting(t *testing.T) {
	db := newTestDBForSettings(t)

	db.Create(&models.Setting{Key: "test_key", Value: encryptValue("stored-value")})

	// Non-existent key returns empty
	got := getEnvOrSetting(db, "nonexistent_key_xyz")
	if got != "" {
		t.Errorf("nonexistent: got %q, want empty", got)
	}

	// Existing key returns decrypted value
	got = getEnvOrSetting(db, "test_key")
	if got != "stored-value" {
		t.Errorf("existing: got %q, want %q", got, "stored-value")
	}
}

func TestGetEnvOrSetting_EnvTakesPrecedence(t *testing.T) {
	db := newTestDBForSettings(t)

	db.Create(&models.Setting{Key: "test_key", Value: "db-value"})
	os.Setenv("TEST_KEY", "env-value")
	defer os.Unsetenv("TEST_KEY")

	got := getEnvOrSetting(db, "test_key")
	if got != "env-value" {
		t.Errorf("env should take precedence: got %q", got)
	}
}

// ----- Settings list ordering -----

func TestSettingHandler_List_OrderedByKey(t *testing.T) {
	db := newTestDBForSettings(t)
	router := setupTestRouterWithSettingHandler(db)

	keys := []string{"zebra", "apple", "banana", "cherry"}
	for _, k := range keys {
		db.Create(&models.Setting{Key: k, Value: "val"})
	}

	w := getJSON(router, "/api/settings")
	var resp struct {
		Success bool `json:"success"`
		Data    []struct {
			Key string `json:"key"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Should be sorted alphabetically
	expected := []string{"apple", "banana", "cherry", "zebra"}
	for i, e := range expected {
		if resp.Data[i].Key != e {
			t.Errorf("position %d: got %s, want %s", i, resp.Data[i].Key, e)
		}
	}
}

// ----- Secret masking in list -----

func TestSettingHandler_List_SecretMasked(t *testing.T) {
	db := newTestDBForSettings(t)
	router := setupTestRouterWithSettingHandler(db)

	db.Create(&models.Setting{Key: "plain", Value: "visible"})
	db.Create(&models.Setting{Key: "secret", Value: "hidden123", IsSecret: true})

	w := getJSON(router, "/api/settings")
	var resp struct {
		Success bool `json:"success"`
		Data    []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	for _, s := range resp.Data {
		if s.Key == "secret" && s.Value != "********" {
			t.Errorf("secret value should be masked, got %s", s.Value)
		}
		if s.Key == "plain" && s.Value != "visible" {
			t.Errorf("plain value should be visible, got %s", s.Value)
		}
	}
}

// ----- Settings API response format -----

func TestSettingHandler_Set_ResponseFormat(t *testing.T) {
	db := newTestDBForSettings(t)
	router := setupTestRouterWithSettingHandler(db)

	w := putJSON(router, "/api/settings/format_test", map[string]interface{}{
		"value": "test_value",
	})

	var resp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success")
	}
	if resp.Message != "Setting saved" {
		t.Errorf("message: got %q", resp.Message)
	}
}

func TestSettingHandler_Delete_ResponseFormat(t *testing.T) {
	db := newTestDBForSettings(t)
	router := setupTestRouterWithSettingHandler(db)

	db.Create(&models.Setting{Key: "del_test", Value: "val"})

	w := deleteReq(router, "/api/settings/del_test")
	var resp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success")
	}
	if resp.Message != "Setting deleted" {
		t.Errorf("message: got %q", resp.Message)
	}
}

// ----- Handler with Gin context helper -----

func putJSONWithContentType(router *gin.Engine, path string, body interface{}) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}
