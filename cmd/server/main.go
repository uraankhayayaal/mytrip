package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	"mytrip/internal/api/handlers"
	"mytrip/internal/api/middleware"
	"mytrip/internal/config"
	"mytrip/internal/domain"
	"mytrip/internal/jwtutil"
	"mytrip/internal/repo"
	"mytrip/internal/service"
)

// authDeps — зависимости аутентификации (CHAT-07: A7-01).
type authDeps struct {
	auth      service.AuthService
	validator jwtutil.TokenValidator
}

// tripDeps — зависимости CRUD поездок (CHAT-07: A7-02).
type tripDeps struct {
	trips service.TripService
}

var deps *authDeps
var trips *tripDeps

func newAuthDeps(cfg *config.Config, pool *pgxpool.Pool) (*authDeps, error) {
	userRepo := repo.NewPGUserRepo(pool)
	tokenRepo := repo.NewPGRefreshTokenRepo(pool)
	issuer, validator, err := jwtutil.NewHS256(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	if err != nil {
		return nil, err
	}
	auth := service.NewAuthService(userRepo, tokenRepo, tokenRepo, issuer, validator, cfg.JWTRefreshTTL)
	return &authDeps{auth: auth, validator: validator}, nil
}

// authMiddleware — валидация Bearer JWT access-токена (CHAT-07: A7-01/A7-05).
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		const prefix = "Bearer "
		if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
			middleware.WriteError(w, domain.NewUnauthorized("missing or malformed Authorization header"))
			return
		}
		claims, err := deps.validator.Validate(header[len(prefix):], jwtutil.TokenKindAccess)
		if err != nil {
			middleware.WriteError(w, domain.NewUnauthorized("invalid or expired token"))
			return
		}
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// pgxpool
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("pgxpool: %v", err)
	}
	defer pool.Close()

	// golang-migrate: source file://db/migrations, database postgres://
	if err := runMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// S3 client (aws-sdk-go-v2, endpoint cfg.S3Endpoint, credentials static)
	s3Client, err := newS3Client(ctx, cfg)
	if err != nil {
		log.Fatalf("s3: %v", err)
	}
	_ = s3Client // TODO(BEL-04): передать в storage-сервис

	// Auth (CHAT-07: A7-01)
	deps, err = newAuthDeps(cfg, pool)
	if err != nil {
		log.Fatalf("auth: %v", err)
	}

	// Trips (CHAT-07: A7-02): TripRepo + TripService + TripHandler под auth-middleware.
	trips = &tripDeps{trips: service.NewTripService(repo.NewPGTripRepo(pool))}

	// chi router
	r := chi.NewRouter()
	r.Use(middleware.ErrorHandler)
	r.Get("/healthz", healthHandler)
	r.Route("/api/v1", func(r chi.Router) {
		handlers.NewAuthHandler(deps.auth).Mount(r)
		// /trips* — только для авторизованных (A7-02).
		r.Use(authMiddleware)
		handlers.NewTripHandler(trips.trips).Mount(r)
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("server listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server: %v", err)
	}
}

func toAppError(err error) *domain.AppError {
	if ae, ok := err.(*domain.AppError); ok {
		return ae
	}
	return domain.NewInternal()
}

// healthHandler — GET /healthz → 200 {"status":"ok"} (без auth).
// Контракт с ARCH-02 (healthcheck compose-сервиса).
func healthHandler(w http.ResponseWriter, r *http.Request) {
	middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func runMigrations(databaseURL string) error {
	source, err := (&file.File{}).Open("file://db/migrations")
	if err != nil {
		return err
	}
	db, err := (&postgres.Postgres{}).Open(databaseURL)
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("file", source, "postgres", db)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func newS3Client(ctx context.Context, cfg *config.Config) (*s3.Client, error) {
	if cfg.S3Endpoint == "" {
		return nil, errors.New("S3_ENDPOINT is empty")
	}
	cred := credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, "")
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithCredentialsProvider(cred))
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.S3Endpoint)
	})
	return client, nil
}
