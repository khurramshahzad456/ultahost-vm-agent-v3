package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	NestAPIBase string
	OpenAIKey   string
}

var AppConfig *Config

func LoadConfig() {
	// Load .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println(" .env file not found, relying on environment variables")
	}

	AppConfig = &Config{
		Port:        getEnv("PORT", "8089"),
		NestAPIBase: getEnv("NEST_API_URL", "https://api.ultahost.dev"),
		OpenAIKey:   getEnv("OPENAI_KEY", ""),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
