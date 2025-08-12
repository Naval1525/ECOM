package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

type DatabaseConfig struct {
    URL      string `mapstructure:"url"`
    Host     string `mapstructure:"host"`
    Port     string `mapstructure:"port"`
    User     string `mapstructure:"user"`
    Password string `mapstructure:"password"`
    Name     string `mapstructure:"name"`
    SSLMode  string `mapstructure:"sslmode"`
}

type ServerConfig struct {
    Port string `mapstructure:"port"`
}

type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    JWTSecret string        `mapstructure:"jwt_secret"`
    Database DatabaseConfig `mapstructure:"database"`
}

func Load() (*Config, error) {
    v := viper.New()
    // Load .env into process env if present (optional)
    _ = gotenv.Load()

    // Config file support: config.yaml/json/toml at project root (optional)
    v.SetConfigName("config")
    v.AddConfigPath(".")
    v.AddConfigPath("./cmd")
    v.AddConfigPath("./configs")
    v.SetConfigType("yaml")
    _ = v.ReadInConfig() // ignore error if not found; env will still work

    // Environment variables override
    v.SetEnvPrefix("")
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    // Explicit env bindings for keys used during Unmarshal
    _ = v.BindEnv("jwt_secret", "JWT_SECRET")
    _ = v.BindEnv("server.port", "SERVER_PORT")
    // Common alias for PaaS
    if v.GetString("server.port") == "" {
        if p := os.Getenv("PORT"); p != "" {
            v.Set("server.port", p)
        }
    }

    // Support both DB_URL and DATABASE_URL aliases
    if v.GetString("database.url") == "" {
        if url := os.Getenv("DB_URL"); url != "" {
            v.Set("database.url", url)
        } else if url := os.Getenv("DATABASE_URL"); url != "" {
            v.Set("database.url", url)
        }
    }
    // Bind discrete database fields for Unmarshal
    _ = v.BindEnv("database.host", "DB_HOST")
    _ = v.BindEnv("database.port", "DB_PORT")
    _ = v.BindEnv("database.user", "DB_USER")
    _ = v.BindEnv("database.password", "DB_PASSWORD")
    _ = v.BindEnv("database.name", "DB_NAME")
    _ = v.BindEnv("database.sslmode", "DB_SSLMODE")
    // Support PORT alias for server.port
    if v.GetString("server.port") == "" {
        if p := v.GetString("PORT"); p != "" {
            v.Set("server.port", p)
        }
    }

    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    return &cfg, nil
}

// Validate ensures required config values are present (no in-code fallbacks)
func (c *Config) Validate() error {
    if c.JWTSecret == "" {
        return fmt.Errorf("JWT secret is required (env JWT_SECRET)")
    }
    if c.Server.Port == "" {
        return fmt.Errorf("server port is required (env SERVER_PORT or PORT)")
    }
    if c.Database.URL == "" {
        // Require discrete fields
        if c.Database.Host == "" || c.Database.Port == "" || c.Database.User == "" || c.Database.Name == "" {
            return fmt.Errorf("database configuration is required (set DATABASE_URL/DB_URL or all of DB_HOST, DB_PORT, DB_USER, DB_NAME)")
        }
    }
    return nil
}


