package routes_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Aryma-f4/uas-backend/config"
	"github.com/Aryma-f4/uas-backend/models"
	"github.com/Aryma-f4/uas-backend/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github. com/stretchr/testify/assert"
	"database/sql"
	"go.mongodb.org/mongo-driver/mongo"
)

var app *fiber.App
var validStudentToken string
var validAdvisorToken string
var validAdminToken string

const testJWTSecret = "test-secret-very-long-and-secure-1234567890"

// TestMain dipanggil sekali sebelum semua test dijalankan
func TestMain(m *testing.M) {
	// RECOVERY: Catch panic di initialization
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[PANIC] TestMain initialization failed: %v\n", r)
		}
	}()

	app = fiber.New()

	// Set JWT_SECRET agar middleware bisa validate token
	os.Setenv("JWT_SECRET", testJWTSecret)

	// Mock DB dan Mongo dengan nil (simulating test environment tanpa koneksi database)
	var db *sql.DB = nil
	var mongoDB *mongo.Database = nil

	// Nil pointer yang bertipe TEPAT *config.Config
	var cfg *config.Config = nil

	fmt.Println("[TEST] Initializing routes with nil databases...")
	
	// Setup routes — sekarang tipe cocok 100% dan tidak panic
	routes.SetupRoutes(app, db, mongoDB, cfg)

	fmt.Println("[TEST] Routes initialized successfully")

	// Generate token valid untuk testing
	validStudentToken = generateTestToken(uuid.New(), uuid.New())
	validAdvisorToken = generateTestToken(uuid.New(), uuid.New())
	validAdminToken = generateTestToken(uuid.New(), uuid.New())

	fmt.Println("[TEST] Generated test tokens successfully")
	fmt.Printf("[TEST] validStudentToken: %s\n", validStudentToken[: 20]+"...")

	// Jalankan semua test
	code := m.Run()
	
	os.Exit(code)
}

// generateTestToken membuat JWT token untuk testing
func generateTestToken(userID, roleID uuid.UUID) string {
	claims := jwt.MapClaims{
		"user_id":  userID. String(),
		"role_id":  roleID. String(),
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testJWTSecret))
	return signed
}

// performRequest menjalankan HTTP request menggunakan app.Test
// Mengembalikan *http.Response untuk divalidasi status code-nya
func performRequest(method, url string, body interface{}, token string) *http.Response {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}

	// Membuat http.Request
	httpReq, _ := http.NewRequest(method, url, &buf)
	httpReq.Header.Set("Content-Type", "application/json")
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	// Menggunakan app.Test untuk mendapatkan response
	resp, err := app.Test(httpReq, -1) // -1 = no timeout
	if err != nil {
		fmt.Printf("[ERROR] app.Test failed: %v\n", err)
		panic(err) // Panic jika ada error pada test request
	}

	return resp
}

// ====================== TEST CASES ======================

func TestAuthLogin_ExpectUnauthorized_NoDB(t *testing.T) {
	payload := models.LoginRequest{
		Username: "anyone",
		Password: "anything",
	}

	resp := performRequest("POST", "/api/v1/auth/login", payload, "")
	defer resp.Body.Close()
	
	// Validasi status code
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 Unauthorized")
}

func TestAuthProfile_WithValidToken(t *testing.T) {
	resp := performRequest("GET", "/api/v1/auth/profile", nil, validStudentToken)
	defer resp.Body.Close()
	
	// Dengan valid token, status tidak harus 401
	assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode, "expected status != 401 with valid token")
}

func TestCreateAchievement_WithValidToken(t *testing.T) {
	payload := models.CreateAchievementRequest{
		AchievementType: "competition",
		Title:           "Test Prestasi",
		Description:     "Deskripsi test",
	}

	resp := performRequest("POST", "/api/v1/achievements", payload, validStudentToken)
	defer resp.Body.Close()
	
	// Dengan valid token, status tidak harus 401
	assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode, "expected status != 401 with valid token")
}

func TestCreateAchievement_WithoutToken(t *testing.T) {
	payload := models.CreateAchievementRequest{
		AchievementType: "competition",
		Title:           "Test",
	}

	resp := performRequest("POST", "/api/v1/achievements", payload, "")
	defer resp.Body.Close()
	
	// Tanpa token, harus 401
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 Unauthorized without token")
}

func TestListAchievements_WithValidToken(t *testing.T) {
	resp := performRequest("GET", "/api/v1/achievements", nil, validStudentToken)
	defer resp.Body.Close()
	
	// Dengan valid token, status tidak harus 401
	assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode, "expected status != 401 with valid token")
}

func TestSubmitAchievement_WithValidToken(t *testing.T) {
	id := uuid.New().String()
	resp := performRequest("POST", "/api/v1/achievements/"+id+"/submit", nil, validStudentToken)
	defer resp.Body.Close()
	
	// Dengan valid token, status tidak harus 401
	assert. NotEqual(t, http.StatusUnauthorized, resp.StatusCode, "expected status != 401 with valid token")
}

func TestListUsers_AsAdminToken(t *testing.T) {
	resp := performRequest("GET", "/api/v1/users", nil, validAdminToken)
	defer resp.Body.Close()
	
	// Dengan valid token, status tidak harus 401
	assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode, "expected status != 401 with valid admin token")
}

func TestListUsers_AsStudentToken(t *testing.T) {
	resp := performRequest("GET", "/api/v1/users", nil, validStudentToken)
	defer resp.Body.Close()
	
	// Dengan valid token, status tidak harus 401
	assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode, "expected status != 401 with valid student token")
}

func TestGetStatistics_WithValidToken(t *testing.T) {
	resp := performRequest("GET", "/api/v1/reports/statistics", nil, validStudentToken)
	defer resp.Body.Close()
	
	// Dengan valid token, status tidak harus 401
	assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode, "expected status != 401 with valid token")
}
