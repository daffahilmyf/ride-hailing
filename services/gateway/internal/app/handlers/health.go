package handlers

import (
	"net/http"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/responses"
	"github.com/gin-gonic/gin"
)

func Health() gin.HandlerFunc {
	return func(c *gin.Context) {
		responses.RespondOK(c, http.StatusOK, map[string]string{"status": "ok"})
	}
}
