package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Server
	Port string

	// PostgreSQL
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresSSL      string

	// MongoDB
	MongoURI string
	MongoDB  string

	// JWT
	JWTSecret          string
	JWTExpireHours     int
	JWTRefreshExpHours int
}

func LoadConfig() *Config {
	jwtExpire, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "24"))
	jwtRefreshExpire, _ := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRE_HOURS", "168"))

	return &Config{
		Port:               getEnv("PORT", "3000"),
		PostgresHost:       getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:       getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:       getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword:   getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:         getEnv("POSTGRES_DB", "achievement_db"),
		PostgresSSL:        getEnv("POSTGRES_SSL", "disable"),
		MongoURI:           getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:            getEnv("MONGO_DB", "achievement_db"),
		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpireHours:     jwtExpire,
		JWTRefreshExpHours: jwtRefreshExpire,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
