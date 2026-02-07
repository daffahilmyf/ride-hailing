package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	locationv1 "github.com/daffahilmyf/ride-hailing/proto/location/v1"
	matchingv1 "github.com/daffahilmyf/ride-hailing/proto/matching/v1"
	grpcadapter "github.com/daffahilmyf/ride-hailing/services/gateway/internal/adapters/grpc"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/handlers/requests"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/handlers/validators"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/responses"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/ports/outbound"
)

func UpdateDriverStatus(matchingClient outbound.MatchingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.UpdateDriverStatusRequest
		if !validators.BindAndValidate(c, &req) {
			responses.RespondErrorCode(c, responses.CodeValidationError, nil)
			return
		}

		driverID := c.Param("driver_id")
		if _, err := uuid.Parse(driverID); err != nil {
			responses.RespondErrorCode(c, responses.CodeValidationError, map[string]string{"field": "driver_id"})
			return
		}

		ctx := grpcadapter.WithRequestMetadata(
			c.Request.Context(),
			contextdata.GetTraceID(c),
			contextdata.GetRequestID(c),
		)
		WithGRPCMeta(c, "matching-service")

		resp, err := matchingClient.UpdateDriverStatus(ctx, &matchingv1.UpdateDriverStatusRequest{
			DriverId:  driverID,
			Status:    req.Status,
			TraceId:   contextdata.GetTraceID(c),
			RequestId: contextdata.GetRequestID(c),
		})
		if err != nil {
			code, details := responses.MapGRPCError(err)
			responses.RespondErrorCode(c, code, details)
			return
		}

		responses.RespondOK(c, 200, map[string]interface{}{
			"status": resp.GetStatus(),
		})
	}
}

func UpdateDriverLocation(locationClient outbound.LocationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.UpdateDriverLocationRequest
		if !validators.BindAndValidate(c, &req) {
			responses.RespondErrorCode(c, responses.CodeValidationError, nil)
			return
		}

		driverID := c.Param("driver_id")
		if _, err := uuid.Parse(driverID); err != nil {
			responses.RespondErrorCode(c, responses.CodeValidationError, map[string]string{"field": "driver_id"})
			return
		}

		ctx := grpcadapter.WithRequestMetadata(
			c.Request.Context(),
			contextdata.GetTraceID(c),
			contextdata.GetRequestID(c),
		)
		WithGRPCMeta(c, "location-service")

		resp, err := locationClient.UpdateDriverLocation(ctx, &locationv1.UpdateDriverLocationRequest{
			DriverId:  driverID,
			Lat:       req.Lat,
			Lng:       req.Lng,
			AccuracyM: req.AccuracyM,
			TraceId:   contextdata.GetTraceID(c),
			RequestId: contextdata.GetRequestID(c),
		})
		if err != nil {
			code, details := responses.MapGRPCError(err)
			responses.RespondErrorCode(c, code, details)
			return
		}

		responses.RespondOK(c, 200, map[string]interface{}{
			"status": resp.GetStatus(),
		})
	}
}
