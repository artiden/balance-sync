package app

import (
	"context"
	"log"
	"sync"

	"goservice/internal/config"
	"goservice/internal/db"
	"goservice/internal/rabbit"
	"goservice/internal/service"

	"gorm.io/gorm"
)

// App holds all application components for coordinated lifecycle management.
type App struct {
	cfg      *config.Config
	db       *gorm.DB
	consumer *rabbit.Consumer
	syncer   *service.Syncer
}

// New creates a new application instance with all dependencies.
func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	mysql, err := db.Connect(cfg)
	if err != nil {
		return nil, err
	}

	cache := service.NewCache()
	repo := service.NewRepository(mysql)
	syncer := service.NewSyncer(cache, repo)

	consumer, err := rabbit.NewConsumer(cfg, service.NewHandler(cache, repo))
	if err != nil {
		return nil, err
	}

	return &App{
		cfg:      cfg,
		db:       mysql,
		consumer: consumer,
		syncer:   syncer,
	}, nil
}

// Run starts all application components and blocks until the context is cancelled.
func (a *App) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	// Start syncer in background
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.syncer.Start(ctx)
	}()

	// Start consumer (blocks until context cancelled or error)
	err := a.consumer.Start(ctx)

	// Wait for syncer to finish
	wg.Wait()

	return err
}

// Shutdown gracefully shuts down all application components.
func (a *App) Shutdown() error {
	log.Printf("app: shutting down...")

	// Close consumer
	if err := a.consumer.Close(); err != nil {
		log.Printf("app: error closing consumer: %v", err)
	}

	// Close database connection
	sqlDB, err := a.db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("app: error closing database: %v", err)
		}
	}

	log.Printf("app: shutdown complete")
	return nil
}

// Run is a convenience function that creates and runs the application.
// Deprecated: Use New() and App.Run() for better control.
func Run(ctx context.Context) error {
	app, err := New()
	if err != nil {
		return err
	}
	return app.Run(ctx)
}
