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
	cfg := config.Load()

	initCtx := context.Background()

	infra, cleanup, err := initInfrastructure(initCtx, cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	useCases := initUseCases(infra)

	srv := setupHTTPServer(cfg, useCases, infra)

	return runServer(srv)
}

type infrastructure struct {
	tokenGenerator token.Generator
	passwordHasher hasher.Hasher
	fileStore      filestore.Store

	userRepo    *authInfra.PgUserRepo
	chatRepo    *chatInfra.PgChatRepo
	messageRepo *chatInfra.PgMessageRepo

	authPortal *authPortal.Portal
}

type useCases struct {
	auth         authuc.UseCase
	user         useruc.UseCase
	chat         chatuc.UseCase
	message      messageuc.UseCase
	notification notificationuc.UseCase
}

func initInfrastructure(ctx context.Context, cfg *config.Config) (*infrastructure, func(), error) {
	pool, err := pg.NewPostgresPool(ctx, cfg.Postgres.DSN())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init postgres pool: %w", err)
	}

	tokenGenerator := token.NewGenerator(
		cfg.AuthToken.Secret,
		cfg.AuthToken.AccessTokenTTL,
		cfg.AuthToken.RefreshTokenTTL,
	)
	passwordHasher := hasher.NewHasher(100000, 16, 32)
	fileStore := filestore.NewMinioStore(filestore.Config{
		Endpoint:        cfg.MinIO.Endpoint,
		Bucket:          cfg.MinIO.Bucket,
		AccessKeyID:     cfg.MinIO.AccessKeyID,
		SecretAccessKey: cfg.MinIO.SecretAccessKey,
		UseSSL:          cfg.MinIO.UseSSL,
	})

	userRepo := authInfra.NewPgUserRepo(pool)
	chatRepo := chatInfra.NewPgChatRepo(pool)
	messageRepo := chatInfra.NewPgMessageRepo(pool)

	authPr := authPortal.New(userRepo, tokenGenerator)

	cleanup := func() {
		pool.Close()
		log.Println("Postgres pool closed")
	}

	return &infrastructure{
		tokenGenerator: tokenGenerator,
		passwordHasher: passwordHasher,
		fileStore:      fileStore,
		userRepo:       userRepo,
		chatRepo:       chatRepo,
		messageRepo:    messageRepo,
		authPortal:     authPr,
	}, cleanup, nil
}

func initUseCases(infra *infrastructure) *useCases {
	return &useCases{
		auth:         authuc.New(infra.userRepo, infra.passwordHasher, infra.tokenGenerator),
		user:         useruc.New(infra.userRepo, infra.passwordHasher, infra.fileStore, infra.authPortal),
		chat:         chatuc.New(infra.chatRepo, infra.messageRepo, infra.authPortal),
		message:      messageuc.New(infra.chatRepo, infra.messageRepo, infra.authPortal),
		notification: notificationuc.New(infra.chatRepo, infra.messageRepo, infra.authPortal),
	}
}

func setupHTTPServer(cfg *config.Config, uc *useCases, infra *infrastructure) *http.Server {
	mux := http.NewServeMux()

	authHttp.Register(mux, "/auth", uc.auth, uc.user, infra.tokenGenerator, infra.authPortal)
	chatHttp.Register(mux, "/chat", uc.chat, uc.message, uc.notification, infra.authPortal)

	return &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
}

func runServer(srv *http.Server) error {
	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Printf("Received shutdown signal: %v", sig)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			if closeErr := srv.Close(); closeErr != nil {
				return fmt.Errorf("failed to close server: %w", closeErr)
			}
			return fmt.Errorf("failed to gracefully shutdown server: %w", err)
		}

		log.Println("Server stopped gracefully")
		return nil
	}
}
