package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/responses"
)

type Readiness struct {
	Redis *redis.Client
	GRPC  []*grpc.ClientConn
	Cache ReadinessCache
}

func Ready(check Readiness) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cached, ok, _ := check.Cache.Get(c.Request.Context()); ok && cached == "ok" {
			responses.RespondOK(c, http.StatusOK, map[string]string{"status": "ready"})
			return
		}

		if check.Redis != nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
			defer cancel()
			if err := check.Redis.Ping(ctx).Err(); err != nil {
				responses.RespondErrorCode(c, responses.CodeInternal, map[string]string{"reason": "REDIS_DOWN"})
				return
			}
		}
		for _, conn := range check.GRPC {
			if conn == nil {
				responses.RespondErrorCode(c, responses.CodeInternal, map[string]string{"reason": "GRPC_CONN_NIL"})
				return
			}
			state := conn.GetState()
			if state == connectivity.TransientFailure || state == connectivity.Shutdown {
				responses.RespondErrorCode(c, responses.CodeInternal, map[string]string{"reason": "GRPC_DOWN"})
				return
			}
		}

		_ = check.Cache.Set(c.Request.Context(), "ok")
		responses.RespondOK(c, http.StatusOK, map[string]string{"status": "ready"})
	}
}
