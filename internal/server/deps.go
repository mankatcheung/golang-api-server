package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	exchangeadapter "github.com/golang-api-server/internal/adapter/exchange"
	"github.com/golang-api-server/internal/cache"
	"github.com/golang-api-server/internal/config"
	"github.com/golang-api-server/internal/consumer"
	"github.com/golang-api-server/internal/database"
	"github.com/golang-api-server/internal/logger"
	"github.com/golang-api-server/internal/repository"
	"github.com/golang-api-server/internal/service"
	"github.com/golang-api-server/pkg/kafka"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type Dependencies struct {
	Config      *config.Config
	DB          *gorm.DB
	TxManager   database.TransactionManager
	Cache       cache.Cache
	UserRepo    service.UserRepository
	UserSvc     service.UserService
	AuthService service.AuthService
	ExchangeSvc service.ExchangeService
	EventSvc    service.EventPublisher
	Producer    service.MessageProducer
	Consumer    consumer.MessageConsumer
	LogProducer service.MessageProducer
	LogConsumer consumer.MessageConsumer
	LogHandler  *logger.KafkaHandler
}

func NewDependencies(cfg *config.Config) (*Dependencies, error) {
	db, err := initDatabase(cfg)
	if err != nil {
		return nil, err
	}

	producer, err := initKafka(cfg.KafkaBrokers, cfg.KafkaTopic)
	if err != nil {
		return nil, err
	}

	logProducer, err := initKafka(cfg.KafkaBrokers, cfg.LogKafkaTopic)
	if err != nil {
		return nil, err
	}

	consumerStart := consumer.StartConversionConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID)
	logConsumer := consumer.StartLogConsumer(cfg.KafkaBrokers, cfg.LogKafkaTopic, cfg.LogKafkaGroupID, db)

	redisCache := cache.NewRedisCache(cfg.RedisAddrs, cfg.RedisPassword, cfg.RedisDB)

	txManager := database.NewTransactionManager(db)
	baseUserRepo := repository.NewUserRepository(db)
	cachedUserRepo := repository.NewCachingUserRepository(baseUserRepo, redisCache, cfg.CacheTTL)
	userService := service.NewUserService(cachedUserRepo)
	authService := newAuthService(cfg, baseUserRepo)

	baseExchangeClient := exchangeadapter.NewHTTPClient(0)
	cachedExchangeClient := exchangeadapter.NewCachedHTTPClient(baseExchangeClient, redisCache, cfg.CacheTTL)
	exchangeService := service.NewExchangeService(cachedExchangeClient)

	eventService := service.NewEventService(producer)

	return &Dependencies{
		Config:      cfg,
		DB:          db,
		TxManager:   txManager,
		Cache:       redisCache,
		UserRepo:    cachedUserRepo,
		UserSvc:     userService,
		AuthService: authService,
		ExchangeSvc: exchangeService,
		EventSvc:    eventService,
		Producer:    producer,
		Consumer:    consumerStart,
		LogProducer: logProducer,
		LogConsumer: logConsumer,
	}, nil
}

func (d *Dependencies) Close() {
	if err := d.Consumer.Close(); err != nil {
		slog.Error("close consumer", "error", err)
	}
	if err := d.LogConsumer.Close(); err != nil {
		slog.Error("close log consumer", "error", err)
	}
	if d.LogHandler != nil {
		d.LogHandler.Close()
	}
	if closer, ok := d.Producer.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			slog.Error("close producer", "error", err)
		}
	}
	if closer, ok := d.LogProducer.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			slog.Error("close log producer", "error", err)
		}
	}
	if c, ok := d.Cache.(interface{ Close() error }); ok {
		if err := c.Close(); err != nil {
			slog.Error("close cache", "error", err)
		}
	}
	sqlDB, err := d.DB.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			slog.Error("close database", "error", err)
		}
	}
}

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if err := db.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{postgres.Open(cfg.DatabaseURL)},
		Replicas: []gorm.Dialector{postgres.Open(cfg.DatabaseReadURL)},
		Policy:   dbresolver.RandomPolicy{},
	})); err != nil {
		return nil, fmt.Errorf("configure db resolver: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying sql.DB: %w", err)
	}

	if err := sqlDB.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	sqlDB.SetMaxOpenConns(int(cfg.DBMaxConns))
	sqlDB.SetMaxIdleConns(int(cfg.DBMinConns))
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	slog.Info("connected to database")

	return db, nil
}

func initKafka(brokers []string, topic string) (service.MessageProducer, error) {
	if err := kafka.EnsureTopic(brokers, topic, 1); err != nil {
		return nil, fmt.Errorf("ensure kafka topic: %w", err)
	}
	return kafka.NewProducer(brokers, topic), nil
}

func newAuthService(cfg *config.Config, userRepo service.UserRepository) service.AuthService {
	return service.NewAuthService(service.AuthServiceDeps{
		UserRepo:      userRepo,
		JWTSecret:     cfg.JWTSecret,
		AccessExpiry:  cfg.JWTAccessExpiry,
		RefreshExpiry: cfg.JWTRefreshExpiry,
	})
}
