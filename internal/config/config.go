package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App       AppConfig       `yaml:"app"`
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	RabbitMQ  RabbitMQConfig  `yaml:"rabbitmq"`
	Worker    WorkerConfig    `yaml:"worker"`
	Ethiopian EthiopianConfig `yaml:"ethiopian"`
	Logging   LoggingConfig   `yaml:"logging"`
}

type AppConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
}

type ServerConfig struct {
	Port                    int           `yaml:"port"`
	ReadTimeout             time.Duration `yaml:"read_timeout"`
	WriteTimeout            time.Duration `yaml:"write_timeout"`
	GracefulShutdownTimeout time.Duration `yaml:"graceful_shutdown_timeout"`
}

type DatabaseConfig struct {
	Host                  string        `yaml:"host"`
	Port                  int           `yaml:"port"`
	User                  string        `yaml:"user"`
	Password              string        `yaml:"password"`
	Name                  string        `yaml:"name"`
	SSLMode               string        `yaml:"sslmode"`
	MaxConnections        int           `yaml:"max_connections"`
	MaxIdleConnections    int           `yaml:"max_idle_connections"`
	ConnectionMaxLifetime time.Duration `yaml:"connection_max_lifetime"`
}

type RabbitMQConfig struct {
	URL           string `yaml:"url"`
	QueueName     string `yaml:"queue_name"`
	Exchange      string `yaml:"exchange"`
	ConsumerTag   string `yaml:"consumer_tag"`
	PrefetchCount int    `yaml:"prefetch_count"`
}

type WorkerConfig struct {
	Concurrency int           `yaml:"concurrency"`
	MaxRetries  int           `yaml:"max_retries"`
	RetryDelay  time.Duration `yaml:"retry_delay"`
}

// Ethiopian-specific configuration
type EthiopianConfig struct {
	USDToETBRate       float64  `yaml:"usd_to_etb"`
	BusinessHoursStart string   `yaml:"business_hours_start"`
	BusinessHoursEnd   string   `yaml:"business_hours_end"`
	ReferencePrefixes  []string `yaml:"reference_prefixes"`
	MaxETBAmount       float64  `yaml:"max_etb_amount"` // Ethiopian regulatory limit
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// Load configuration from YAML and environment variables
func Load() (*Config, error) {
	// Try to load from YAML first
	cfg := &Config{}

	// Read config file
	data, err := os.ReadFile("config.yaml")
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config.yaml: %w", err)
		}
	}

	// Override with environment variables
	overrideFromEnv(cfg)

	return cfg, nil
}

func overrideFromEnv(cfg *Config) {
	// Database
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Database.Port = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Database.Password = password
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		cfg.Database.Name = name
	}

	// RabbitMQ
	if url := os.Getenv("RABBITMQ_URL"); url != "" {
		cfg.RabbitMQ.URL = url
	}
	if queue := os.Getenv("RABBITMQ_QUEUE"); queue != "" {
		cfg.RabbitMQ.QueueName = queue
	}

	// Server
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	}

	// Worker
	if concurrency := os.Getenv("WORKER_CONCURRENCY"); concurrency != "" {
		if c, err := strconv.Atoi(concurrency); err == nil {
			cfg.Worker.Concurrency = c
		}
	}

	// Ethiopian
	if rate := os.Getenv("ETB_USD_RATE"); rate != "" {
		if r, err := strconv.ParseFloat(rate, 64); err == nil {
			cfg.Ethiopian.USDToETBRate = r
		}
	}
}
