package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	ridev1 "github.com/daffahilmyf/ride-hailing/proto/ride/v1"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
)

type fakeRideClient struct{}

type captureRideClient struct {
	lastCreate *ridev1.CreateRideRequest
	lastCancel *ridev1.CancelRideRequest
	lastOffer  *ridev1.CreateOfferRequest
	lastAction *ridev1.OfferActionRequest
}

func (f *captureRideClient) CreateRide(ctx context.Context, in *ridev1.CreateRideRequest, opts ...grpc.CallOption) (*ridev1.CreateRideResponse, error) {
	f.lastCreate = in
	return &ridev1.CreateRideResponse{RideId: "r1", Status: "MATCHING"}, nil
}

func (f *captureRideClient) CancelRide(ctx context.Context, in *ridev1.CancelRideRequest, opts ...grpc.CallOption) (*ridev1.CancelRideResponse, error) {
	f.lastCancel = in
	return &ridev1.CancelRideResponse{RideId: in.RideId, Status: "CANCELLED"}, nil
}

func (f *captureRideClient) CreateOffer(ctx context.Context, in *ridev1.CreateOfferRequest, opts ...grpc.CallOption) (*ridev1.CreateOfferResponse, error) {
	f.lastOffer = in
	return &ridev1.CreateOfferResponse{OfferId: "o1", RideId: in.RideId, DriverId: in.DriverId, Status: "PENDING"}, nil
}

func (f *captureRideClient) AcceptOffer(ctx context.Context, in *ridev1.OfferActionRequest, opts ...grpc.CallOption) (*ridev1.OfferActionResponse, error) {
	f.lastAction = in
	return &ridev1.OfferActionResponse{OfferId: in.OfferId, RideId: "r1", DriverId: "d1", Status: "ACCEPTED"}, nil
}

func (f *captureRideClient) DeclineOffer(ctx context.Context, in *ridev1.OfferActionRequest, opts ...grpc.CallOption) (*ridev1.OfferActionResponse, error) {
	f.lastAction = in
	return &ridev1.OfferActionResponse{OfferId: in.OfferId, RideId: "r1", DriverId: "d1", Status: "DECLINED"}, nil
}

func (f *captureRideClient) ExpireOffer(ctx context.Context, in *ridev1.OfferActionRequest, opts ...grpc.CallOption) (*ridev1.OfferActionResponse, error) {
	f.lastAction = in
	return &ridev1.OfferActionResponse{OfferId: in.OfferId, RideId: "r1", DriverId: "d1", Status: "EXPIRED"}, nil
}

func setupRideRouter(client *captureRideClient, withUser bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if traceID := c.GetHeader("X-Trace-Id"); traceID != "" {
			contextdata.SetTraceID(c, traceID)
		}
		if requestID := c.GetHeader("X-Request-Id"); requestID != "" {
			contextdata.SetRequestID(c, requestID)
		}
		c.Next()
	})
	if withUser {
		r.Use(func(c *gin.Context) {
			contextdata.SetUserContext(c, "11111111-1111-1111-1111-111111111111", "rider")
			c.Next()
		})
	}
	r.POST("/rides", CreateRide(client, ""))
	r.POST("/rides/:ride_id/cancel", CancelRide(client, ""))
	r.POST("/rides/:ride_id/offers", CreateOffer(client, ""))
	r.POST("/offers/:offer_id/accept", AcceptOffer(client, ""))
	r.POST("/offers/:offer_id/decline", DeclineOffer(client, ""))
	r.POST("/offers/:offer_id/expire", ExpireOffer(client, ""))
	return r
}

func TestCreateRideValidation(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		status int
	}{
		{"missing_fields", `{"pickup_lat":1}`, http.StatusBadRequest},
		{"ok", `{"pickup_lat":1,"pickup_lng":2,"dropoff_lat":3,"dropoff_lng":4}`, http.StatusOK},
	}

	client := &captureRideClient{}
	r := setupRideRouter(client, true)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/rides", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, w.Code)
			}
		})
	}
}

func TestCreateRideAuthRequired(t *testing.T) {
	client := &captureRideClient{}
	r := setupRideRouter(client, false)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/rides", bytes.NewBufferString(`{"pickup_lat":1,"pickup_lng":2,"dropoff_lat":3,"dropoff_lng":4}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestCreateRideMetadataHeaders(t *testing.T) {
	client := &captureRideClient{}
	r := setupRideRouter(client, true)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/rides", bytes.NewBufferString(`{"pickup_lat":1,"pickup_lng":2,"dropoff_lat":3,"dropoff_lng":4}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-1")
	req.Header.Set("X-Trace-Id", "trace-1")
	req.Header.Set("X-Request-Id", "req-1")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
	if client.lastCreate == nil {
		t.Fatalf("expected create request")
	}
	if client.lastCreate.GetIdempotencyKey() != "idem-1" {
		t.Fatalf("expected idempotency key")
	}
	if client.lastCreate.GetTraceId() != "trace-1" {
		t.Fatalf("expected trace id")
	}
	if client.lastCreate.GetRequestId() != "req-1" {
		t.Fatalf("expected request id")
	}
}

func TestCancelRideValidation(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		body   string
		status int
	}{
		{"bad_id", "/rides/123/cancel", `{"reason":"r"}`, http.StatusBadRequest},
		{"missing_reason", "/rides/11111111-1111-1111-1111-111111111111/cancel", `{"reason":""}`, http.StatusBadRequest},
		{"ok", "/rides/11111111-1111-1111-1111-111111111111/cancel", `{"reason":"r"}`, http.StatusOK},
	}

	client := &captureRideClient{}
	r := setupRideRouter(client, true)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, w.Code)
			}
		})
	}
}

func TestCreateOfferValidation(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		body   string
		status int
	}{
		{"bad_ride_id", "/rides/123/offers", `{"driver_id":"11111111-1111-1111-1111-111111111111"}`, http.StatusBadRequest},
		{"missing_driver_id", "/rides/11111111-1111-1111-1111-111111111111/offers", `{}`, http.StatusBadRequest},
		{"bad_driver_id", "/rides/11111111-1111-1111-1111-111111111111/offers", `{"driver_id":"bad"}`, http.StatusBadRequest},
		{"ok", "/rides/11111111-1111-1111-1111-111111111111/offers", `{"driver_id":"11111111-1111-1111-1111-111111111111","offer_ttl_seconds":10}`, http.StatusOK},
	}

	client := &captureRideClient{}
	r := setupRideRouter(client, true)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, w.Code)
			}
		})
	}
}

func TestOfferActionValidation(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		status int
	}{
		{"bad_id", "/offers/123/accept", http.StatusBadRequest},
		{"ok", "/offers/11111111-1111-1111-1111-111111111111/accept", http.StatusOK},
	}

	client := &captureRideClient{}
	r := setupRideRouter(client, true)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", tt.path, bytes.NewBufferString(`{}`))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, w.Code)
			}
		})
	}
}
