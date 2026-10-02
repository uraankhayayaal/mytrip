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

	"mytrip/internal/api/middleware"
	"mytrip/internal/config"
	"mytrip/internal/domain"
)

// Точки расширения DI: конструкторы сервисов/репозиториев/хендлеров
// добавят BEL-03..06. Пока nil — маршруты /api/v1/* появятся вместе с ними.
var (
// TODO(BEL-03): tripService *service.TripService — инициализировать в main
// TODO(BEL-04): photoService *service.PhotoService
// TODO(BEL-05): stopService *service.StopService
// TODO(BEL-06): authService *service.AuthService
)

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

	// chi router
	r := chi.NewRouter()
	r.Use(middleware.ErrorHandler)
	r.Get("/healthz", healthHandler)
	r.Route("/api/v1", func(r chi.Router) {
		// TODO(BEL-03..06): маршруты /trips, /stops, /photos, /auth
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

var _ = domain.NewInternal // placeholder: domain-контракт используется в middleware
