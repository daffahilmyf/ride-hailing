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
)

type fakeRideClient struct{}

func (f *fakeRideClient) CreateRide(ctx context.Context, in *ridev1.CreateRideRequest, opts ...grpc.CallOption) (*ridev1.CreateRideResponse, error) {
	return &ridev1.CreateRideResponse{RideId: "r1", Status: "MATCHING"}, nil
}

func (f *fakeRideClient) CancelRide(ctx context.Context, in *ridev1.CancelRideRequest, opts ...grpc.CallOption) (*ridev1.CancelRideResponse, error) {
	return &ridev1.CancelRideResponse{RideId: in.RideId, Status: "CANCELLED"}, nil
}

func (f *fakeRideClient) CreateOffer(ctx context.Context, in *ridev1.CreateOfferRequest, opts ...grpc.CallOption) (*ridev1.CreateOfferResponse, error) {
	return &ridev1.CreateOfferResponse{OfferId: "o1", RideId: in.RideId, DriverId: in.DriverId, Status: "PENDING"}, nil
}

func (f *fakeRideClient) AcceptOffer(ctx context.Context, in *ridev1.OfferActionRequest, opts ...grpc.CallOption) (*ridev1.OfferActionResponse, error) {
	return &ridev1.OfferActionResponse{OfferId: in.OfferId, RideId: "r1", DriverId: "d1", Status: "ACCEPTED"}, nil
}

func (f *fakeRideClient) DeclineOffer(ctx context.Context, in *ridev1.OfferActionRequest, opts ...grpc.CallOption) (*ridev1.OfferActionResponse, error) {
	return &ridev1.OfferActionResponse{OfferId: in.OfferId, RideId: "r1", DriverId: "d1", Status: "DECLINED"}, nil
}

func (f *fakeRideClient) ExpireOffer(ctx context.Context, in *ridev1.OfferActionRequest, opts ...grpc.CallOption) (*ridev1.OfferActionResponse, error) {
	return &ridev1.OfferActionResponse{OfferId: in.OfferId, RideId: "r1", DriverId: "d1", Status: "EXPIRED"}, nil
}

func setupRideRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/rides", CreateRide(&fakeRideClient{}))
	r.POST("/rides/:ride_id/cancel", CancelRide(&fakeRideClient{}))
	r.POST("/rides/:ride_id/offers", CreateOffer(&fakeRideClient{}))
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

	r := setupRideRouter()
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

	r := setupRideRouter()
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

	r := setupRideRouter()
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
