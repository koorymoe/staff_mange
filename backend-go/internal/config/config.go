package config

import (
	"os"
	"strconv"
	"strings"
)

// Config يحمل كل إعدادات التطبيق المقروءة من متغيرات البيئة (ENV)
type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	CORSOrigins    []string
	R2Bucket       string
	R2AccessKey    string
	R2SecretKey    string
	R2Endpoint     string
	GeminiAPIKey   string
	GeminiDailyCap int
	TutorialsDir   string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "4000"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		CORSOrigins:    splitOrigins(getEnv("CORS_ORIGIN", "http://localhost:5173")),
		R2Bucket:       getEnv("R2_BUCKET", ""),
		R2AccessKey:    getEnv("R2_ACCESS_KEY", ""),
		R2SecretKey:    getEnv("R2_SECRET_KEY", ""),
		R2Endpoint:     getEnv("R2_ENDPOINT", ""),
		GeminiAPIKey:   getEnv("GEMINI_API_KEY", ""),
		GeminiDailyCap: getEnvInt("GEMINI_DAILY_CAP", 300),
		// TutorialsDir: مسار مجلد أدلة الاستخدام (tutorial-*.html و guide-*.txt) اللي
		// يقرأها المساعد الذكي حتى يعرف يشرح للموظف شلون يستخدم النظام فعلياً.
		// الافتراضي "../tutorials" يناسب التطوير المحلي (تشغيل go run من داخل backend-go
		// حيث مجلد tutorials أخو backend-go بجذر المستودع). بالإنتاج نضبط
		// TUTORIALS_DIR=/app/tutorials صراحة بـ docker-compose.yml مع bind-mount
		// (شوف backend service بـ docker-compose.yml).
		TutorialsDir: getEnv("TUTORIALS_DIR", "../tutorials"),
	}
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func splitOrigins(v string) []string {
	parts := strings.Split(v, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			origins = append(origins, p)
		}
	}
	return origins
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
