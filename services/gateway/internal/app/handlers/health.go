package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/responses"
)

func Health() gin.HandlerFunc {
	return func(c *gin.Context) {
		responses.RespondOK(c, http.StatusOK, map[string]string{"status": "ok"})
	}
}
