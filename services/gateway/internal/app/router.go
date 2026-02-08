package app

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	globalLimiter := cache.NewRedisLimiter(redisClient,
		cache.WithLimiterRequests(cfg.RateLimit.Requests),
		cache.WithLimiterWindow(time.Duration(cfg.RateLimit.WindowSeconds)*time.Second),
		cache.WithLimiterPrefix("gateway:rl"),
	)
	nearbyLimiter := cache.NewRedisLimiter(redisClient,
		cache.WithLimiterRequests(cfg.RateLimit.NearbyRequests),
		cache.WithLimiterWindow(time.Duration(cfg.RateLimit.NearbyWindowSeconds)*time.Second),
		cache.WithLimiterPrefix("gateway:nearby"),
	)
	notifyLimiter := cache.NewRedisLimiter(redisClient,
		cache.WithLimiterRequests(cfg.RateLimit.NotifyRequests),
		cache.WithLimiterWindow(time.Duration(cfg.RateLimit.NotifyWindowSeconds)*time.Second),
		cache.WithLimiterPrefix("gateway:notify"),
	)

	metrics := middleware.NewMetrics(cfg.ServiceName)
	r.Use(middleware.LoggerMiddleware(logger, cfg.ServiceName))
	r.Use(middleware.TraceMiddleware(cfg.ServiceName))
	r.Use(middleware.MetricsMiddleware(metrics))
	r.Use(middleware.RateLimitMiddleware(globalLimiter, cfg.RateLimit.Requests))
	r.Use(middleware.MaxBodyBytes(cfg.MaxBodyBytes))
	r.Use(middleware.RequestTimeout(time.Duration(cfg.HTTP.RequestTimeoutSeconds) * time.Second))

	r.StaticFile("/favicon.ico", "config/favicon.ico")
	r.GET("/openapi.yaml", func(c *gin.Context) {
		paths := []string{"/config/openapi.yaml", "config/openapi.yaml"}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				c.File(path)
				return
			}
		}
		c.Status(http.StatusNotFound)
	})
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/openapi.yaml")))
	var grpcConns []*grpc.ClientConn
	if grpcClients != nil {
		grpcConns = []*grpc.ClientConn{grpcClients.RideConn, grpcClients.MatchingConn, grpcClients.LocationConn, grpcClients.UserConn}
	}

	readyCache := handlers.ReadinessCache{
		Cache: cache.NewRedisCache(redisClient),
		TTL:   cache.DefaultTTL(cfg.Cache),
		Key:   "gateway:readyz",
	}

	r.GET("/readyz", handlers.Ready(handlers.Readiness{
		Redis: redisClient,
		GRPC:  grpcConns,
		Cache: readyCache,
	}))

	v1 := r.Group("/v1")
	{
		v1.GET("/healthz", handlers.Health())
		v1.GET("/readyz", handlers.Ready(handlers.Readiness{
			Redis: redisClient,
			GRPC:  grpcConns,
			Cache: readyCache,
		}))
		v1.POST("/auth/register", handlers.RegisterAuth(deps.AuthClient, cfg.GRPC.InternalToken))
		v1.POST("/auth/login", handlers.LoginAuth(deps.AuthClient, cfg.GRPC.InternalToken))
		v1.POST("/auth/refresh", handlers.RefreshAuth(deps.AuthClient, cfg.GRPC.InternalToken))
		v1.POST("/auth/logout", handlers.LogoutAuth(deps.AuthClient, cfg.GRPC.InternalToken))
		v1.POST("/auth/verify", handlers.VerifyAuth(deps.AuthClient, cfg.GRPC.InternalToken))

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
		riderGroup.POST("/rides", handlers.CreateRide(deps.RideClient, cfg.GRPC.InternalToken))
		riderGroup.POST("/rides/:ride_id/cancel", handlers.CancelRide(deps.RideClient, cfg.GRPC.InternalToken))
		riderGroup.POST("/rides/:ride_id/offers", handlers.CreateOffer(deps.RideClient, cfg.GRPC.InternalToken))

		driverGroup := authGroup.Group("/")
		driverGroup.Use(middleware.RequireRole(middleware.RoleDriver))
		driverGroup.Use(middleware.RequireScope("drivers:write"))
		driverGroup.Use(middleware.AuditLogger(logger, "drivers:write"))
		driverGroup.POST("/drivers/status", handlers.UpdateDriverStatus(deps.MatchingClient))
		driverGroup.POST("/drivers/location", handlers.UpdateDriverLocation(deps.LocationClient, cfg.GRPC.InternalToken))
		driverGroup.POST("/drivers/nearby",
			middleware.RateLimitMiddleware(nearbyLimiter, cfg.RateLimit.NearbyRequests),
			handlers.ListNearbyDrivers(deps.LocationClient, cfg.GRPC.InternalToken),
		)
		driverGroup.POST("/offers/:offer_id/accept", handlers.AcceptOffer(deps.RideClient, cfg.GRPC.InternalToken))
		driverGroup.POST("/offers/:offer_id/decline", handlers.DeclineOffer(deps.RideClient, cfg.GRPC.InternalToken))
		driverGroup.POST("/offers/:offer_id/expire", handlers.ExpireOffer(deps.RideClient, cfg.GRPC.InternalToken))

		adminGroup := authGroup.Group("/admin")
		adminGroup.Use(middleware.RequireRole(middleware.RoleAdmin))
		adminGroup.Use(middleware.RequireScope("admin:drivers:write"))
		adminGroup.Use(middleware.AuditLogger(logger, "admin:drivers:write"))
		adminGroup.POST("/drivers/:driver_id/status", handlers.UpdateDriverStatusFor(deps.MatchingClient))
		adminGroup.POST("/drivers/:driver_id/location", handlers.UpdateDriverLocationFor(deps.LocationClient, cfg.GRPC.InternalToken))

		userGroup := authGroup.Group("/")
		userGroup.Use(middleware.RequireRole(middleware.RoleRider, middleware.RoleDriver))
		userGroup.GET("/users/me", handlers.MeAuth(deps.AuthClient, cfg.GRPC.InternalToken))
		userGroup.POST("/auth/logout_all", handlers.LogoutAllAuth(deps.AuthClient, cfg.GRPC.InternalToken))
		userGroup.POST("/auth/logout_device", handlers.LogoutDeviceAuth(deps.AuthClient, cfg.GRPC.InternalToken))
		userGroup.GET("/auth/sessions", handlers.ListSessionsAuth(deps.AuthClient, cfg.GRPC.InternalToken))

		authGroup.GET("/notify/sse",
			middleware.RateLimitMiddleware(notifyLimiter, cfg.RateLimit.NotifyRequests),
			middleware.RequireScope("notify:read"),
			handlers.StreamNotifications(cfg.Notify.BaseURL),
		)
	}

	return r
}
