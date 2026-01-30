package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"goservice/internal/app"
)

func main() {
	// Create context that cancels on SIGINT or SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create application
	application, err := app.New()
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	// Run shutdown handler in background
	go func() {
		sig := <-sigChan
		log.Printf("received signal: %v, initiating shutdown...", sig)
		cancel()
		application.Shutdown()
	}()

	log.Printf("starting service...")
	if err := application.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("service failed: %v", err)
	}

	log.Printf("service stopped")
}
