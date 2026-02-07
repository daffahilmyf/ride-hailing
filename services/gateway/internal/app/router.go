package app

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/adapters/cache"
	grpcadapter "github.com/daffahilmyf/ride-hailing/services/gateway/internal/adapters/grpc"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/handlers"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/middleware"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/infra"
)

func NewRouter(cfg infra.Config, logger *zap.Logger, deps Deps, redisClient *redis.Client, grpcClients *grpcadapter.Clients) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware(logger, cfg.ServiceName))
	r.Use(middleware.RequestTimeout(time.Duration(cfg.HTTP.RequestTimeoutSeconds) * time.Second))
	if cfg.Observability.TracingEnabled {
		r.Use(middleware.TraceMiddleware(cfg.ServiceName))
	}
	if cfg.Observability.MetricsEnabled {
		metrics := middleware.NewMetrics(cfg.ServiceName)
		registry := prometheus.NewRegistry()
		registry.MustRegister(metrics.Requests, metrics.Latency)
		r.Use(middleware.MetricsMiddleware(metrics))
		r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
	}
	r.Use(middleware.MaxBodyBytes(cfg.MaxBodyBytes))

	limiter := cache.NewRedisLimiter(
		redisClient,
		cache.WithLimiterRequests(cfg.RateLimit.Requests),
		cache.WithLimiterWindow(time.Duration(cfg.RateLimit.WindowSeconds)*time.Second),
		cache.WithLimiterPrefix("rl"),
	)
	r.Use(middleware.RateLimitMiddleware(limiter, cfg.RateLimit.Requests))

	var readyCache = handlers.ReadinessCache{Key: "gateway:readyz"}
	if cfg.Cache.Enabled {
		readyCache.Cache = cache.NewRedisCache(redisClient)
		readyCache.TTL = cache.DefaultTTL(cfg.Cache)
	}

	r.GET("/healthz", handlers.Health())
	r.GET("/readyz", handlers.Ready(handlers.Readiness{
		Redis: redisClient,
		GRPC:  []*grpc.ClientConn{grpcClients.RideConn, grpcClients.MatchingConn, grpcClients.LocationConn},
		Cache: readyCache,
	}))

	v1 := r.Group("/v1")
	{
		v1.GET("/healthz", handlers.Health())
		v1.GET("/readyz", handlers.Ready(handlers.Readiness{
			Redis: redisClient,
			GRPC:  []*grpc.ClientConn{grpcClients.RideConn, grpcClients.MatchingConn, grpcClients.LocationConn},
			Cache: readyCache,
		}))

		authGroup := v1.Group("/")
		authGroup.Use(middleware.AuthMiddleware(logger, middleware.AuthConfig{
			Enabled:   cfg.Auth.Enabled,
			JWTSecret: cfg.Auth.JWTSecret,
			Issuer:    cfg.Auth.Issuer,
			Audience:  cfg.Auth.Audience,
		}))

		riderGroup := authGroup.Group("/")
		riderGroup.Use(middleware.RequireRole(middleware.RoleRider))
		riderGroup.Use(middleware.RequireScope("rides:write"))
		riderGroup.Use(middleware.AuditLogger(logger, "rides:write"))
		riderGroup.POST("/rides", handlers.CreateRide(deps.RideClient))
		riderGroup.POST("/rides/:ride_id/cancel", handlers.CancelRide(deps.RideClient))

		driverGroup := authGroup.Group("/")
		driverGroup.Use(middleware.RequireRole(middleware.RoleDriver))
		driverGroup.Use(middleware.RequireScope("drivers:write"))
		driverGroup.Use(middleware.AuditLogger(logger, "drivers:write"))
		driverGroup.POST("/drivers/:driver_id/status", handlers.UpdateDriverStatus(deps.MatchingClient))
		driverGroup.POST("/drivers/:driver_id/location", handlers.UpdateDriverLocation(deps.LocationClient))
	}

	return r
}
