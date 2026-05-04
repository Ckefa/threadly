package configs

import "os"

type Config struct {
	Port string
}

func Load() *Config {
	return &Config{
		Port: get("APP_PORT", "8080"),
	}
}

func get(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
