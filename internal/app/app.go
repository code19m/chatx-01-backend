package app

import (
	"bufio"
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
	"chatx-01-backend/pkg/middleware"
	"chatx-01-backend/pkg/pg"
	"chatx-01-backend/pkg/token"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	cfg   *config.Config
	pool  *pgxpool.Pool
	infra *infrastructure
	uc    *useCases
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

func Build(ctx context.Context) (*App, error) {
	cfg := config.Load()

	pool, err := pg.NewPostgresPool(ctx, cfg.Postgres.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to init postgres pool: %w", err)
	}

	infra := initInfrastructure(pool, cfg)
	uc := initUseCases(infra)

	return &App{
		cfg:   cfg,
		pool:  pool,
		infra: infra,
		uc:    uc,
	}, nil
}

func (a *App) Close() {
	if a.pool != nil {
		a.pool.Close()
		log.Println("Postgres pool closed")
	}
}

func initInfrastructure(pool *pgxpool.Pool, cfg *config.Config) *infrastructure {
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

	return &infrastructure{
		tokenGenerator: tokenGenerator,
		passwordHasher: passwordHasher,
		fileStore:      fileStore,
		userRepo:       userRepo,
		chatRepo:       chatRepo,
		messageRepo:    messageRepo,
		authPortal:     authPr,
	}
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

func (a *App) RunHTTPServer() error {
	srv := a.setupHTTPServer()
	return a.runServer(srv)
}

func (a *App) setupHTTPServer() *http.Server {
	// base handler/router/server
	mux := http.NewServeMux()

	// register module handlers
	authHttp.Register(mux, "/auth", a.uc.auth, a.uc.user, a.infra.tokenGenerator, a.infra.authPortal)
	chatHttp.Register(mux, "/chat", a.uc.chat, a.uc.message, a.uc.notification, a.infra.authPortal)

	// global middlewares
	handler := middleware.CORS(mux)

	return &http.Server{
		Addr:         a.cfg.Server.Addr,
		Handler:      handler,
		ReadTimeout:  a.cfg.Server.ReadTimeout,
		WriteTimeout: a.cfg.Server.WriteTimeout,
		IdleTimeout:  a.cfg.Server.IdleTimeout,
	}
}

func (a *App) runServer(srv *http.Server) error {
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

func (a *App) CreateSuperUser() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	email = strings.TrimSpace(email)

	fmt.Print("Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	username = strings.TrimSpace(username)

	fmt.Print("Password: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	password = strings.TrimSpace(password)

	fmt.Print("Password (confirm): ")
	passwordConfirm, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	passwordConfirm = strings.TrimSpace(passwordConfirm)

	if password != passwordConfirm {
		return fmt.Errorf("passwords do not match")
	}

	req := useruc.CreateSuperUserReq{
		Email:    email,
		Username: username,
		Password: password,
	}

	if err := req.Validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	resp, err := a.uc.user.CreateSuperUser(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to create super user: %w", err)
	}

	fmt.Printf("\nSuper user created successfully!\n")
	fmt.Printf("User ID: %d\n", resp.UserID)
	fmt.Printf("Email: %s\n", email)
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Role: admin\n")

	return nil
}
