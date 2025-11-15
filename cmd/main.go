package main

import (
	authHttp "chatx-01-backend/internal/auth/controller/http"
	authInfra "chatx-01-backend/internal/auth/infra"
	authPortal "chatx-01-backend/internal/auth/portal"
	"chatx-01-backend/internal/auth/usecase/authuc"
	"chatx-01-backend/internal/auth/usecase/useruc"
	chatHttp "chatx-01-backend/internal/chat/controller/http"
	chatInfra "chatx-01-backend/internal/chat/infra"
	"chatx-01-backend/internal/chat/usecase/chatuc"
	"chatx-01-backend/internal/chat/usecase/messageuc"
	"chatx-01-backend/internal/chat/usecase/notificationuc"
	"chatx-01-backend/internal/config"
	"chatx-01-backend/pkg/filestore"
	"chatx-01-backend/pkg/hasher"
	"chatx-01-backend/pkg/pg"
	"chatx-01-backend/pkg/token"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	initCtx := context.Background()

	// Load configuration from environment variables
	cfg := config.Load()

	// Initialize database pool
	pool, err := pg.NewPostgresPool(initCtx, cfg.Postgres.DSN())
	if err != nil {
		return fmt.Errorf("failed to init postgres pool: %w", err)
	}
	defer pool.Close()

	// Init pkg components
	tokenGenerator := token.NewGenerator(
		cfg.AuthToken.Secret,
		cfg.AuthToken.AccessTokenTTL,
		cfg.AuthToken.RefreshTokenTTL,
	)
	passwordHasher := hasher.NewHasher(100000, 16, 32) // PBKDF2 with 100k iterations, 16-byte salt, 32-byte key
	fileStore := filestore.NewMinioStore(filestore.Config{
		Endpoint:        cfg.MinIO.Endpoint,
		Bucket:          cfg.MinIO.Bucket,
		AccessKeyID:     cfg.MinIO.AccessKeyID,
		SecretAccessKey: cfg.MinIO.SecretAccessKey,
		UseSSL:          cfg.MinIO.UseSSL,
	})

	// Initialize repositories
	pgUser := authInfra.NewPgUserRepo(pool)
	pgChat := chatInfra.NewPgChatRepo(pool)
	pgMessage := chatInfra.NewPgMessageRepo(pool)

	// Initialize portals
	authPr := authPortal.New(pgUser, tokenGenerator)

	// Initialize use cases
	authUseCase := authuc.New(pgUser, passwordHasher, tokenGenerator)
	userUseCase := useruc.New(pgUser, passwordHasher, fileStore, authPr)
	chatUseCase := chatuc.New(pgChat, pgMessage, authPr)
	messageUseCase := messageuc.New(pgChat, pgMessage, authPr)
	notificationUseCase := notificationuc.New(pgChat, pgMessage, authPr)

	// Initialize mux server
	mux := http.NewServeMux()

	// Register controllers
	authHttp.Register(mux, "/auth", authUseCase, userUseCase, tokenGenerator, authPr)
	chatHttp.Register(mux, "/chat", chatUseCase, messageUseCase, notificationUseCase, authPr)

	// Configure HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Channel to listen for errors from the server
	serverErrors := make(chan error, 1)

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("Starting server on %s", cfg.Server.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Printf("Received shutdown signal: %v", sig)

		// Create context with timeout for graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := srv.Shutdown(shutdownCtx); err != nil {
			// Force close if graceful shutdown fails
			if closeErr := srv.Close(); closeErr != nil {
				return fmt.Errorf("failed to close server: %w", closeErr)
			}
			return fmt.Errorf("failed to gracefully shutdown server: %w", err)
		}

		log.Println("Server stopped gracefully")
		return nil
	}
}
