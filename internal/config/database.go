package config

import "os"

type DBConfig struct {
	Host        string
	Port        string
	Username    string
	Password    string
	Database    string
	PostgresDSN string
}

func NewDBConfig() DBConfig {
	return DBConfig{
		Host:        os.Getenv("DB_HOST"),
		Port:        os.Getenv("DB_PORT"),
		Username:    os.Getenv("DB_USERNAME"),
		Password:    os.Getenv("DB_PASSWORD"),
		Database:    os.Getenv("DB_DATABASE"),
		PostgresDSN: os.Getenv("POSTGRES_DSN"),
	}
}
