package main

import (
	"log/slog"
	"sync"
	"time"

	"github.com/hayohtee/go-backend-template/internal/data"
	"github.com/hayohtee/go-backend-template/internal/mailer"
)

// application holds the dependencies for the HTTP handlers
// helpers and middlewares.
type application struct {
	config config
	logger *slog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

// config struct holds the configuration settings
// for the application.
type config struct {
	// the port to listen on
	port int
	// the current operating environment (development|staging|production...)
	env string
	// the configuration settings for database connection pool.
	db struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	// the configuration settings for rate limit
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	// the configurations settings for smtp
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}
