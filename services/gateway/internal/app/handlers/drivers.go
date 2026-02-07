package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/handlers/requests"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/handlers/validators"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/responses"
)

func UpdateDriverStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.UpdateDriverStatusRequest
		if !validators.BindAndValidate(c, &req) {
			responses.RespondErrorCode(c, responses.CodeValidationError, nil)
			return
		}
		// gRPC call placeholder:
		// if err := driverClient.UpdateStatus(...); err != nil {
		//   code, details := responses.MapGRPCError(err)
		//   responses.RespondErrorCode(c, code, details)
		//   return
		// }
		responses.RespondNotImplemented(c)
	}
}

func UpdateDriverLocation() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.UpdateDriverLocationRequest
		if !validators.BindAndValidate(c, &req) {
			responses.RespondErrorCode(c, responses.CodeValidationError, nil)
			return
		}
		// gRPC call placeholder:
		// if err := locationClient.UpdateDriverLocation(...); err != nil {
		//   code, details := responses.MapGRPCError(err)
		//   responses.RespondErrorCode(c, code, details)
		//   return
		// }
		responses.RespondNotImplemented(c)
	}
}
