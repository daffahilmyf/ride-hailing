package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/handlers/requests"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/handlers/validators"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/responses"
)

func CreateRide() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.CreateRideRequest
		if !validators.BindAndValidate(c, &req) {
			responses.RespondErrorCode(c, responses.CodeValidationError, nil)
			return
		}
		responses.RespondNotImplemented(c)
	}
}

func CancelRide() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.CancelRideRequest
		if !validators.BindAndValidate(c, &req) {
			responses.RespondErrorCode(c, responses.CodeValidationError, nil)
			return
		}
		responses.RespondNotImplemented(c)
	}
}
