package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/boopateway/boopateway/internal/framework/booprecognition"
	"github.com/boopateway/boopateway/internal/framework/config"
	"github.com/boopateway/boopateway/internal/framework/database/sqlite"
	"github.com/boopateway/boopateway/internal/framework/httpserver"
	"github.com/boopateway/boopateway/internal/framework/httpserver/handler"
	"github.com/boopateway/boopateway/internal/framework/security"
	"github.com/boopateway/boopateway/internal/framework/storage"
	"github.com/boopateway/boopateway/internal/usecase"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// framework
	db, err := sqlite.Open(ctx, cfg.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	images, err := storage.NewLocalImageStorage(cfg.ImageDir)
	if err != nil {
		return err
	}
	recognition := booprecognition.NewFaceRecognitionGateway(booprecognition.NewClient(booprecognition.Config{
		BaseURL:    cfg.BoopBaseURL,
		Token:      cfg.BoopToken,
		SearchPath: cfg.BoopSearchPath,
	}))
	hasher := security.BcryptCodeHasher{}
	persons := sqlite.NewPersonRepository(db)
	transactions := sqlite.NewTransactionRepository(db)
	idempotency := sqlite.NewIdempotencyRepository(db)

	// use cases
	createPerson := usecase.NewCreatePersonUseCase(persons, recognition, images, hasher)
	getPerson := usecase.NewGetPersonUseCase(persons)
	createPayment := usecase.NewCreatePaymentUseCase(persons, transactions, recognition)
	getPayment := usecase.NewGetPaymentUseCase(transactions)
	confirmPayment := usecase.NewConfirmPaymentUseCase(persons, transactions, hasher)

	// http
	router := httpserver.NewRouter(httpserver.Handlers{
		Person:  handler.NewPersonHandler(createPerson, getPerson),
		Payment: handler.NewPaymentHandler(createPayment, getPayment, confirmPayment),
	}, idempotency)

	srv := &http.Server{Addr: cfg.Addr, Handler: router, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	slog.Info("listening", "addr", cfg.Addr, "boop", cfg.BoopBaseURL)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
