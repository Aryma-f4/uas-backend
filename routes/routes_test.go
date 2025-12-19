package routes_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Aryma-f4/uas-backend/config" // <-- IMPORTANT: import ini
	"github.com/Aryma-f4/uas-backend/models"
	"github.com/Aryma-f4/uas-backend/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"database/sql"
	"go.mongodb.org/mongo-driver/mongo"
)

var app *fiber.App
var validStudentToken string
var validAdvisorToken string
var validAdminToken string

const testJWTSecret = "test-secret-very-long-and-secure-1234567890"

func TestMain(m *testing.M) {
	app = fiber.New()

	// Set JWT_SECRET agar middleware bisa validate token
	os.Setenv("JWT_SECRET", testJWTSecret)

	// Mock DB dan Mongo dengan nil
	var db *sql.DB = nil
	var mongoDB *mongo.Database = nil

	// Nil pointer yang bertipe TEPAT *config.Config
	var cfg *config.Config = nil

	// Setup routes — sekarang tipe cocok 100%
	routes.SetupRoutes(app, db, mongoDB, cfg)

	// Generate token valid
	validStudentToken = generateTestToken(uuid.New(), uuid.New())
	validAdvisorToken = generateTestToken(uuid.New(), uuid.New())
	validAdminToken = generateTestToken(uuid.New(), uuid.New())

	m.Run()
}

func generateTestToken(userID, roleID uuid.UUID) string {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role_id": roleID.String(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testJWTSecret))
	return signed
}

func performRequest(method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}

	req := httptest.NewRequest(method, url, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer " + token)
	}

	rr := httptest.NewRecorder()
	app.Test(req)

	return rr
}

// ====================== TEST CASES ======================

func TestAuthLogin_ExpectUnauthorized_NoDB(t *testing.T) {
	payload := models.LoginRequest{
		Username: "anyone",
		Password: "anything",
	}

	rr := performRequest("POST", "/api/v1/auth/login", payload, "")
	assert.Equal(t, 401, rr.Code)
}

func TestAuthProfile_WithValidToken(t *testing.T) {
	rr := performRequest("GET", "/api/v1/auth/profile", nil, validStudentToken)
	assert.NotEqual(t, 401, rr.Code)
}

func TestCreateAchievement_WithValidToken(t *testing.T) {
	payload := models.CreateAchievementRequest{
		AchievementType: "competition",
		Title:           "Test Prestasi",
		Description:     "Deskripsi test",
	}

	rr := performRequest("POST", "/api/v1/achievements", payload, validStudentToken)
	assert.NotEqual(t, 401, rr.Code)
}

func TestCreateAchievement_WithoutToken(t *testing.T) {
	payload := models.CreateAchievementRequest{
		AchievementType: "competition",
		Title:           "Test",
	}

	rr := performRequest("POST", "/api/v1/achievements", payload, "")
	assert.Equal(t, 401, rr.Code)
}

func TestListAchievements_WithValidToken(t *testing.T) {
	rr := performRequest("GET", "/api/v1/achievements", nil, validStudentToken)
	assert.NotEqual(t, 401, rr.Code)
}

func TestSubmitAchievement_WithValidToken(t *testing.T) {
	id := uuid.New().String()
	rr := performRequest("POST", "/api/v1/achievements/"+id+"/submit", nil, validStudentToken)
	assert.NotEqual(t, 401, rr.Code)
}

func TestListUsers_AsAdminToken(t *testing.T) {
	rr := performRequest("GET", "/api/v1/users", nil, validAdminToken)
	assert.NotEqual(t, 401, rr.Code)
}

func TestListUsers_AsStudentToken(t *testing.T) {
	rr := performRequest("GET", "/api/v1/users", nil, validStudentToken)
	assert.NotEqual(t, 401, rr.Code)
}

func TestGetStatistics_WithValidToken(t *testing.T) {
	rr := performRequest("GET", "/api/v1/reports/statistics", nil, validStudentToken)
	assert.NotEqual(t, 401, rr.Code)
}
