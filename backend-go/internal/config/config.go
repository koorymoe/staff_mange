package config

import (
	"os"
)

// Config يحمل كل إعدادات التطبيق المقروءة من متغيرات البيئة (ENV)
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	R2Bucket    string
	R2AccessKey string
	R2SecretKey string
	R2Endpoint  string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "4000"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		R2Bucket:    getEnv("R2_BUCKET", ""),
		R2AccessKey: getEnv("R2_ACCESS_KEY", ""),
		R2SecretKey: getEnv("R2_SECRET_KEY", ""),
		R2Endpoint:  getEnv("R2_ENDPOINT", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
