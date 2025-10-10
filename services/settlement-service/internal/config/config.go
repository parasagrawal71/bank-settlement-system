package config

import (
	"fmt"

	"github.com/parasagrawal71/bank-settlement-system/shared/env"
)

type Config struct {
	DBUrl    string
	GRPCPort string
}

type DBConfig struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	SSLMode    string
}

func LoadDbConfigFromEnv() DBConfig {
	return DBConfig{
		DBHost:     env.GetEnvString("SETTLEMENT_DB_HOST", ""),
		DBUser:     env.GetEnvString("SETTLEMENT_DB_USER", ""),
		DBPassword: env.GetEnvString("SETTLEMENT_DB_PASSWORD", ""),
		DBName:     env.GetEnvString("SETTLEMENT_DB_NAME", ""),
		DBPort:     env.GetEnvString("SETTLEMENT_DB_PORT", ""),
		SSLMode:    env.GetEnvString("SSL_MODE", "disable"),
	}
}

func Load() *Config {
	dbConfig := LoadDbConfigFromEnv()
	db := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", dbConfig.DBUser, dbConfig.DBPassword, dbConfig.DBHost, dbConfig.DBPort, dbConfig.DBName, dbConfig.SSLMode)

	port := env.GetEnvString("SETTLEMENT_GRPC_PORT", "")
	return &Config{DBUrl: db, GRPCPort: port}
}
