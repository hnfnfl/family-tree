package config

import (
	"fmt"
	"os"
)

type Config struct {
	Server ServerConfig
	Neo4j  Neo4jConfig
	JWT    JWTConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type Neo4jConfig struct {
	URI      string
	Username string
	Password string
	Database string
}

type JWTConfig struct {
	Secret     string
	ExpireHour int
}

func Load() (*Config, error) {
	env := getEnv("APP_ENV", "development")

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Env:  env,
		},
		Neo4j: Neo4jConfig{
			URI:      getEnv("NEO4J_URI", "bolt://localhost:7687"),
			Username: getEnv("NEO4J_USERNAME", "neo4j"),
			Password: getEnv("NEO4J_PASSWORD", "KeluargaTree2026!"),
			Database: getEnv("NEO4J_DATABASE", "neo4j"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			ExpireHour: 24,
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) String() string {
	return fmt.Sprintf("Config{Server.Port: %s, Neo4j.URI: %s, JWT.ExpireHour: %d}",
		c.Server.Port, c.Neo4j.URI, c.JWT.ExpireHour)
}
