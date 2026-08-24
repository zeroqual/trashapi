package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	cfg := Config{
		Server:   loadServerConfig(),
		Database: loadDatabaseConfig(),
	}
	return &cfg
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("DATABASE_HOST", ""),
		Port:     getEnv("DATABASE_PORT", ""),
		User:     getEnv("DATABASE_USER", ""),
		Password: getEnv("DATABASE_PASSWORD", ""),
		Name:     getEnv("DATABASE_NAME", ""),
	}
}

func loadServerConfig() ServerConfig {
	return ServerConfig{
		Host: getEnv("SERVER_HOST", ""),
		Port: getEnv("SERVER_PORT", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c Config) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.Name)
}
