package app

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/handlers"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/middleware"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/infra"
)

func NewRouter(cfg infra.Config, logger *zap.Logger) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware(logger, cfg.ServiceName))

	r.GET("/healthz", handlers.Health())

	v1 := r.Group("/v1")
	{
		v1.GET("/healthz", handlers.Health())

		authGroup := v1.Group("/")
		authGroup.Use(middleware.AuthMiddleware(logger, middleware.AuthConfig{
			Enabled:   cfg.Auth.Enabled,
			JWTSecret: cfg.Auth.JWTSecret,
			Issuer:    cfg.Auth.Issuer,
			Audience:  cfg.Auth.Audience,
		}))

		riderGroup := authGroup.Group("/")
		riderGroup.Use(middleware.RequireRole(middleware.RoleRider))
		riderGroup.POST("/rides", handlers.CreateRide())
		riderGroup.POST("/rides/:ride_id/cancel", handlers.CancelRide())

		driverGroup := authGroup.Group("/")
		driverGroup.Use(middleware.RequireRole(middleware.RoleDriver))
		driverGroup.POST("/drivers/:driver_id/status", handlers.UpdateDriverStatus())
		driverGroup.POST("/drivers/:driver_id/location", handlers.UpdateDriverLocation())
	}

	return r
}
