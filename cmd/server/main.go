package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.tepseg.com/ai/kakao-relay/internal/config"
	"gitlab.tepseg.com/ai/kakao-relay/internal/database"
	"gitlab.tepseg.com/ai/kakao-relay/internal/handler"
	"gitlab.tepseg.com/ai/kakao-relay/internal/jobs"
	"gitlab.tepseg.com/ai/kakao-relay/internal/middleware"
	"gitlab.tepseg.com/ai/kakao-relay/internal/redis"
	"gitlab.tepseg.com/ai/kakao-relay/internal/repository"
	"gitlab.tepseg.com/ai/kakao-relay/internal/service"
	"gitlab.tepseg.com/ai/kakao-relay/internal/sse"
	"gitlab.tepseg.com/ai/kakao-relay/web"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	setLogLevel(cfg.LogLevel)

	isProduction := os.Getenv("K_SERVICE") != "" || os.Getenv("FLY_APP_NAME") != ""
	if err := cfg.Validate(isProduction); err != nil {
		log.Fatal().Err(err).Msg("invalid configuration")
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.DBPingTimeout)
	if err := db.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to ping database")
	}
	cancel()
	log.Info().Msg("database connected")

	redisClient, err := redis.NewClient(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer redisClient.Close()
	log.Info().Msg("redis connected")

	accountRepo := repository.NewAccountRepository(db.DB)
	convRepo := repository.NewConversationRepository(db.DB)
	inboundMsgRepo := repository.NewInboundMessageRepository(db.DB)
	outboundMsgRepo := repository.NewOutboundMessageRepository(db.DB)
	sessionRepo := repository.NewSessionRepository(db.DB)

	broker := sse.NewBroker(redisClient)
	defer broker.Close()

	convService := service.NewConversationService(convRepo)
	messageService := service.NewMessageService(inboundMsgRepo, outboundMsgRepo)
	kakaoService := service.NewKakaoService()
	sessionService := service.NewSessionService(db, sessionRepo, accountRepo, broker)
	ipRateLimiter := service.NewRateLimiter(redisClient.Client)

	authMiddleware := middleware.NewAuthMiddleware(accountRepo, sessionRepo)
	rateLimitMiddleware := middleware.NewRedisRateLimitMiddleware(redisClient.Client)
	kakaoSignatureMiddleware := middleware.NewKakaoSignatureMiddleware(cfg.KakaoSignatureSecret)
	sessionCreateRateLimit := middleware.NewIPRateLimitMiddleware(ipRateLimiter, 10, 5*time.Minute, "session_create")
	sessionStatusRateLimit := middleware.NewIPRateLimitMiddleware(ipRateLimiter, 30, 1*time.Minute, "session_status")
	bodyLimitMiddleware := middleware.NewBodyLimitMiddleware(0)

	kakaoHandler := handler.NewKakaoHandler(
		convService, sessionService, messageService, broker, cfg.CallbackTTL(),
	)
	eventsHandler := handler.NewEventsHandler(broker, messageService)
	openclawHandler := handler.NewOpenClawHandler(messageService, kakaoService)
	sessionHandler := handler.NewSessionHandler(sessionService)

	dashboardRepo := repository.NewDashboardRepository(db.DB)
	dashboardHandler := handler.NewDashboardHandler(
		dashboardRepo, accountRepo, convRepo,
		inboundMsgRepo, outboundMsgRepo,
		sessionService, messageService, broker,
		web.DashboardHTML,
	)

	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestLogger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(config.ServerRequestTimeout))
	r.Use(bodyLimitMiddleware.Handler)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"status":    "ok",
			"timestamp": time.Now().UnixMilli(),
		})
	})

	r.Route("/kakao-talkchannel", func(r chi.Router) {
		r.Use(kakaoSignatureMiddleware.Handler)
		r.Post("/webhook", kakaoHandler.Webhook)
	})

	r.Route("/v1", func(r chi.Router) {
		r.Use(authMiddleware.Handler)
		r.Use(rateLimitMiddleware.Handler)
		r.Get("/events", eventsHandler.ServeHTTP)
	})

	r.Route("/openclaw", func(r chi.Router) {
		r.Use(authMiddleware.Handler)
		r.Use(rateLimitMiddleware.Handler)
		r.Mount("/", openclawHandler.Routes())
	})

	r.Route("/v1/sessions", func(r chi.Router) {
		r.With(sessionCreateRateLimit.Handler).Post("/create", sessionHandler.CreateSession)
		r.With(sessionStatusRateLimit.Handler).Get("/{sessionToken}/status", sessionHandler.GetSessionStatus)
	})

	r.Route("/dashboard", func(r chi.Router) {
		r.Get("/", dashboardHandler.ServeIndex)
		r.Route("/api", func(r chi.Router) {
			r.Get("/overview", dashboardHandler.Overview)
			r.Get("/accounts", dashboardHandler.ListAccounts)
			r.Get("/accounts/{id}/conversations", dashboardHandler.AccountConversations)
			r.Get("/accounts/{id}/messages", dashboardHandler.AccountMessages)
			r.Get("/accounts/{id}/stats", dashboardHandler.AccountStats)
			r.Get("/accounts/{id}/failed-messages", dashboardHandler.AccountFailedMessages)
			r.Post("/accounts/{id}/regenerate-token", dashboardHandler.RegenerateToken)
			r.Delete("/accounts/{id}", dashboardHandler.DeleteAccount)
			r.Delete("/accounts/{id}/conversations/{convId}", dashboardHandler.DeleteConversation)
			r.Get("/sessions", dashboardHandler.ListSessions)
			r.Post("/sessions/create", dashboardHandler.CreateSession)
			r.Post("/sessions/{id}/disconnect", dashboardHandler.DisconnectSession)
			r.Delete("/sessions/{id}", dashboardHandler.DeleteSession)
		})
	})

	cleanupJob := jobs.NewCleanupJob(
		inboundMsgRepo, sessionRepo, config.CleanupJobInterval,
	)
	cleanupJob.Start()
	defer cleanupJob.Stop()

	server := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      r,
		ReadTimeout:  config.ServerReadTimeout,
		WriteTimeout: 0,
		IdleTimeout:  config.ServerIdleTimeout,
	}

	go func() {
		log.Info().Str("addr", cfg.Addr()).Msg("starting server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutting down server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.ServerShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server stopped")
}

func setLogLevel(level string) {
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
