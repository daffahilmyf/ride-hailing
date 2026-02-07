package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	ridev1 "github.com/daffahilmyf/ride-hailing/proto/ride/v1"
	grpcadapter "github.com/daffahilmyf/ride-hailing/services/gateway/internal/adapters/grpc"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/handlers/requests"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/handlers/validators"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/responses"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/ports/outbound"
)

func CreateRide(rideClient outbound.RideService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.CreateRideRequest
		if !validators.BindAndValidate(c, &req) {
			responses.RespondErrorCode(c, responses.CodeValidationError, nil)
			return
		}

		userID := contextdata.GetUserID(c)
		if userID == "" {
			responses.RespondErrorCode(c, responses.CodeUnauthorized, map[string]string{"reason": "MISSING_USER"})
			return
		}

		ctx := grpcadapter.WithRequestMetadata(
			c.Request.Context(),
			contextdata.GetTraceID(c),
			contextdata.GetRequestID(c),
		)
		WithGRPCMeta(c, "ride-service")

		idempotencyKey := c.GetHeader("Idempotency-Key")
		resp, err := rideClient.CreateRide(ctx, &ridev1.CreateRideRequest{
			RiderId:        userID,
			PickupLat:      req.PickupLat,
			PickupLng:      req.PickupLng,
			DropoffLat:     req.DropoffLat,
			DropoffLng:     req.DropoffLng,
			IdempotencyKey: idempotencyKey,
			TraceId:        contextdata.GetTraceID(c),
			RequestId:      contextdata.GetRequestID(c),
		})
		if err != nil {
			code, details := responses.MapGRPCError(err)
			responses.RespondErrorCode(c, code, details)
			return
		}

		responses.RespondOK(c, 200, map[string]interface{}{
			"ride_id": resp.GetRideId(),
			"status":  resp.GetStatus(),
		})
	}
}

func CancelRide(rideClient outbound.RideService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.CancelRideRequest
		if !validators.BindAndValidate(c, &req) {
			responses.RespondErrorCode(c, responses.CodeValidationError, nil)
			return
		}

		rideID := c.Param("ride_id")
		if _, err := uuid.Parse(rideID); err != nil {
			responses.RespondErrorCode(c, responses.CodeValidationError, map[string]string{"field": "ride_id"})
			return
		}

		ctx := grpcadapter.WithRequestMetadata(
			c.Request.Context(),
			contextdata.GetTraceID(c),
			contextdata.GetRequestID(c),
		)
		WithGRPCMeta(c, "ride-service")

		resp, err := rideClient.CancelRide(ctx, &ridev1.CancelRideRequest{
			RideId:    rideID,
			Reason:    req.Reason,
			TraceId:   contextdata.GetTraceID(c),
			RequestId: contextdata.GetRequestID(c),
		})
		if err != nil {
			code, details := responses.MapGRPCError(err)
			responses.RespondErrorCode(c, code, details)
			return
		}

		responses.RespondOK(c, 200, map[string]interface{}{
			"ride_id": resp.GetRideId(),
			"status":  resp.GetStatus(),
		})
	}
}
