package application

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/CatalinPlesu/channel-service/messaging"
	"github.com/CatalinPlesu/channel-service/repository/jwts"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type App struct {
	router   http.Handler
	rdb      *redis.Client
	db       *bun.DB
	rabbitMQ *messaging.RabbitMQ
	config   Config
}

func New(config Config) *App {

	rdb := redis.NewClient(&redis.Options{
		Addr: config.RedisAddress,
	})

	dsn := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=disable", config.PostgresUser, config.PostgresPassword, config.PostgresAddress, config.PostgresDB)
	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqlDB, pgdialect.New())

	rabbitMQ, _ := messaging.NewRabbitMQ(config.RabitMQURL)

	app := &App{
		rdb:      rdb,
		db:       db,
		rabbitMQ: rabbitMQ,
		config:   config,
	}

	app.loadRoutes()

	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.ServerPort),
		Handler: a.router,
	}

	err := a.rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	err = a.db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	go func() {
		if err := a.StartRabbitMQListener(ctx); err != nil {
			fmt.Println("RabbitMQ listener error:", err)
		}
	}()

	defer func() {
		if err := a.rdb.Close(); err != nil {
			fmt.Println("failed to close redis", err)
		}
		if err := a.db.Close(); err != nil {
			fmt.Println("failed to close database", err)
		}
		if err := a.rabbitMQ.Close(); err != nil {
			fmt.Println("failed to close RabbitMQ", err)
		}
	}()

	fmt.Println("Starting server")

	ch := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	select {
	case err = <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		return server.Shutdown(timeout)
	}
}

func (a *App) StartRabbitMQListener(ctx context.Context) error {
	queueName := "user_id_jwt"

	RdRepo := &jwts.RedisRepo{
		Client: a.rdb,
	}

	handler := func(message messaging.LoginRegisterMessage) {

		claims, err := jwts.ValidateJWT(message.JWT)
		if err != nil {
			fmt.Println("bad jwt jwt:", err)
			return
		}

		if claims.UserID != uuid.UUID(message.UserID) {
			fmt.Println("bad jwt no acces or expierd")
			return
		}

		err = RdRepo.Insert(ctx, message.UserID, message.JWT)
		if err != nil {
			fmt.Println("failed to insert user jwt:", err)
			return
		}

		fmt.Printf("Received message: UserID=%s, JWT=%s\n", message.UserID, message.JWT)
	}

	err := a.rabbitMQ.ConsumeLoginRegisterMessages(queueName, handler)
	if err != nil {
		return fmt.Errorf("failed to start RabbitMQ consumer: %w", err)
	}

	<-ctx.Done()
	return nil
}
