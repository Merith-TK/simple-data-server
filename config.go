package main

import (
	"os"
	"strconv"
)

type Config struct {
	Port           string
	DataDir        string
	LogLevel       string
	AllowedOrigins []string
}

func loadConfig() *Config {
	config := &Config{
		Port:           getEnv("PORT", "8080"),
		DataDir:        getEnv("DATA_DIR", "./data"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		AllowedOrigins: []string{"*"}, // TODO: Make this configurable
	}
	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
