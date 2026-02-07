package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/responses"
)

func UpdateDriverStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		responses.RespondNotImplemented(c)
	}
}

func UpdateDriverLocation() gin.HandlerFunc {
	return func(c *gin.Context) {
		responses.RespondNotImplemented(c)
	}
}
