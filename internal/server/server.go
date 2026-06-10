// Package server configures and runs the HTTP and gRPC servers with all dependencies.
//
// It wires together the database, repositories, services, HTTP handlers, gRPC
// handlers, and Kafka components. Both HTTP and gRPC servers start concurrently
// and shut down gracefully on SIGINT/SIGTERM.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-api-server/internal/config"
	grpchandler "github.com/golang-api-server/internal/grpc"
	"github.com/golang-api-server/internal/handler"
	"github.com/golang-api-server/internal/middleware"
	pb "github.com/golang-api-server/proto/auth"
	pbev "github.com/golang-api-server/proto/event"
	pbex "github.com/golang-api-server/proto/exchange"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// App holds all server components and their dependencies.
type App struct {
	deps *Dependencies
	http *http.Server
	grpc *grpclib.Server
}

// New constructs an App from pre-built dependencies.
func New(deps *Dependencies) *App {
	httpHandlers := newHTTPHandlers(deps)
	grpcServer := newGRPCServer(deps)

	httpSrv := &http.Server{
		Addr:         deps.Config.ServerAddress,
		Handler:      setupRouter(deps.Config, httpHandlers),
		ReadTimeout:  deps.Config.ServerTimeout,
		WriteTimeout: deps.Config.ServerTimeout,
		IdleTimeout:  deps.Config.ServerTimeout * 2,
	}

	return &App{
		deps: deps,
		http: httpSrv,
		grpc: grpcServer,
	}
}

// Run starts the HTTP and gRPC servers, blocks until an interrupt signal or a
// startup error, and then shuts everything down gracefully.
func (a *App) Run() error {
	errCh := make(chan error, 2)
	a.startGRPC(errCh)
	a.startHTTP(errCh)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
	case err := <-errCh:
		slog.Error("server startup failed", "error", err)
		a.deps.Close()
		return fmt.Errorf("startup: %w", err)
	}

	return a.shutdown()
}

func (a *App) startGRPC(errCh chan<- error) {
	go func() {
		slog.Info("gRPC server starting", "addr", a.deps.Config.GRPCAddress)
		lis, err := net.Listen("tcp", a.deps.Config.GRPCAddress)
		if err != nil {
			errCh <- fmt.Errorf("gRPC listen: %w", err)
			return
		}
		if err := a.grpc.Serve(lis); err != nil {
			errCh <- fmt.Errorf("gRPC serve: %w", err)
		}
	}()
}

func (a *App) startHTTP(errCh chan<- error) {
	go func() {
		slog.Info("HTTP server starting", "addr", a.deps.Config.ServerAddress)
		if err := a.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("HTTP listen: %w", err)
		}
	}()
}

func (a *App) shutdown() error {
	slog.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var shutdownErr error
	if err := a.http.Shutdown(ctx); err != nil {
		shutdownErr = fmt.Errorf("HTTP shutdown: %w", err)
	}

	a.grpc.GracefulStop()
	a.deps.Close()

	slog.Info("server exited")
	return shutdownErr
}

// ── Handler and gRPC wiring ────────────────────────────────────────────────

// httpHandlers groups all HTTP handlers together.
type httpHandlers struct {
	auth     *handler.AuthHandler
	user     *handler.UserHandler
	exchange *handler.ExchangeHandler
	event    *handler.EventHandler
}

func newHTTPHandlers(deps *Dependencies) *httpHandlers {
	return &httpHandlers{
		auth:     handler.NewAuthHandler(deps.AuthService),
		user:     handler.NewUserHandler(deps.UserSvc),
		exchange: handler.NewExchangeHandler(deps.ExchangeSvc),
		event:    handler.NewEventHandler(deps.EventSvc),
	}
}

func newGRPCServer(deps *Dependencies) *grpclib.Server {
	s := grpclib.NewServer()

	pb.RegisterAuthServiceServer(s, grpchandler.NewAuthGRPCHandler(deps.AuthService))
	pbex.RegisterExchangeServiceServer(s, grpchandler.NewExchangeGRPCHandler(deps.ExchangeSvc))
	pbev.RegisterEventServiceServer(s, grpchandler.NewEventGRPCHandler(deps.EventSvc))

	reflection.Register(s)
	return s
}

// ── Router ─────────────────────────────────────────────────────────────────

func setupRouter(cfg *config.Config, h *httpHandlers) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(corsMiddleware(cfg.AllowedOrigins))
	r.Use(middleware.RequestID())
	r.Use(middleware.Metrics())
	r.Use(middleware.RequestLogger())
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	authLimiter := middleware.NewRateLimiter(ratePerMin(cfg.AuthRateLimit), cfg.AuthRateLimit)
	apiLimiter := middleware.NewRateLimiter(ratePerMin(cfg.APIRateLimit), cfg.APIRateLimit)

	v1 := r.Group("/api/v1")

	auth := v1.Group("/auth")
	auth.Use(authLimiter.Limit())
	{
		auth.POST("/register", h.auth.Register)
		auth.POST("/login", h.auth.Login)
		auth.POST("/refresh", h.auth.RefreshToken)
		auth.POST("/logout", middleware.AuthMiddleware(cfg.JWTSecret), h.auth.Logout)
	}

	protected := v1.Group("/")
	protected.Use(apiLimiter.Limit())
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		protected.GET("/me", h.auth.GetProfile)
		protected.PUT("/me", h.auth.UpdateProfile)
	}

	exchange := v1.Group("/exchange")
	exchange.Use(apiLimiter.Limit())
	exchange.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		exchange.POST("/convert", h.exchange.Convert)
	}

	admin := v1.Group("/admin")
	admin.Use(apiLimiter.Limit())
	admin.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	admin.Use(middleware.RequireRole("admin"))
	{
		admin.GET("/users", h.user.List)
		admin.DELETE("/users/:id", h.user.Delete)
	}

	events := v1.Group("/events")
	events.Use(apiLimiter.Limit())
	events.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		events.POST("/publish", h.event.PublishEvent)
	}

	return r
}

// ratePerMin converts a requests-per-minute integer into a rate.Limit (events/sec).
func ratePerMin(rpm int) rate.Limit {
	if rpm <= 0 {
		return rate.Inf
	}
	return rate.Limit(float64(rpm) / 60.0)
}

func corsMiddleware(allowedOrigins string) gin.HandlerFunc {
	cfg := cors.DefaultConfig()
	if allowedOrigins == "*" {
		cfg.AllowAllOrigins = true
	} else {
		cfg.AllowOrigins = strings.Split(allowedOrigins, ",")
	}
	cfg.AllowHeaders = append(cfg.AllowHeaders, "Authorization")
	return cors.New(cfg)
}
