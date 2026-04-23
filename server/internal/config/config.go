package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	RedisHost string
	RedisPort int

	JWTSecret  string
	ServerPort string

	AllowedOrigins []string
}

func Load() Config {
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		panic("JWT_SECRET environment variable is required")
	}
	if len(jwtSecret) < 32 {
		panic(fmt.Sprintf("JWT_SECRET must be at least 32 characters, got %d", len(jwtSecret)))
	}

	return Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "mmorpg"),

		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnvInt("REDIS_PORT", 6379),

		JWTSecret:  jwtSecret,
		ServerPort: getEnv("SERVER_PORT", ":50051"),

		AllowedOrigins: getEnvSlice("ALLOWED_ORIGINS", []string{"http://localhost:*"}),
	}
}

func (c Config) DSN() string {
	return "host=" + c.DBHost +
		" port=" + strconv.Itoa(c.DBPort) +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=disable"
}

func (c Config) RedisAddr() string {
	return c.RedisHost + ":" + strconv.Itoa(c.RedisPort)
}

func (c Config) IsOriginAllowed(origin string) bool {
	for _, allowed := range c.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
		// Support wildcard patterns like "http://localhost:*"
		if matchWildcard(allowed, origin) {
			return true
		}
	}
	return false
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	var result []string
	for _, s := range splitByComma(value) {
		if s != "" {
			result = append(result, s)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

func splitByComma(s string) []string {
	var parts []string
	current := ""
	for _, c := range s {
		if c == ',' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	parts = append(parts, current)
	return parts
}

func matchWildcard(pattern, s string) bool {
	if pattern == "" {
		return s == ""
	}
	pi := 0
	si := 0
	for pi < len(pattern) && si < len(s) {
		if pattern[pi] == '*' {
			rest := pattern[pi+1:]
			for si <= len(s) {
				if matchWildcard(rest, s[si:]) {
					return true
				}
				si++
			}
			return false
		}
		if pattern[pi] != s[si] {
			return false
		}
		pi++
		si++
	}
	if pi < len(pattern) {
		for pi < len(pattern) {
			if pattern[pi] != '*' {
				return false
			}
			pi++
		}
	}
	return si == len(s)
}
