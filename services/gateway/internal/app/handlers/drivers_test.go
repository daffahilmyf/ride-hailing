package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	locationv1 "github.com/daffahilmyf/ride-hailing/proto/location/v1"
	matchingv1 "github.com/daffahilmyf/ride-hailing/proto/matching/v1"
)

type fakeMatchingClient struct{}

type fakeLocationClient struct{}

func (f *fakeMatchingClient) UpdateDriverStatus(ctx context.Context, in *matchingv1.UpdateDriverStatusRequest, opts ...grpc.CallOption) (*matchingv1.UpdateDriverStatusResponse, error) {
	return &matchingv1.UpdateDriverStatusResponse{Status: "OK"}, nil
}

func (f *fakeLocationClient) UpdateDriverLocation(ctx context.Context, in *locationv1.UpdateDriverLocationRequest, opts ...grpc.CallOption) (*locationv1.UpdateDriverLocationResponse, error) {
	return &locationv1.UpdateDriverLocationResponse{Status: "OK"}, nil
}

func (f *fakeLocationClient) ListNearbyDrivers(ctx context.Context, in *locationv1.ListNearbyDriversRequest, opts ...grpc.CallOption) (*locationv1.ListNearbyDriversResponse, error) {
	return &locationv1.ListNearbyDriversResponse{}, nil
}

func setupDriverRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/drivers/:driver_id/status", UpdateDriverStatus(&fakeMatchingClient{}))
	r.POST("/drivers/:driver_id/location", UpdateDriverLocation(&fakeLocationClient{}, ""))
	return r
}

func TestUpdateDriverStatusValidation(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		body   string
		status int
	}{
		{"bad_id", "/drivers/123/status", `{"status":"ONLINE_AVAILABLE"}`, http.StatusBadRequest},
		{"bad_status", "/drivers/11111111-1111-1111-1111-111111111111/status", `{"status":"BUSY"}`, http.StatusBadRequest},
		{"ok", "/drivers/11111111-1111-1111-1111-111111111111/status", `{"status":"ONLINE_AVAILABLE"}`, http.StatusOK},
	}

	r := setupDriverRouter()
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

func TestUpdateDriverLocationValidation(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		body   string
		status int
	}{
		{"bad_id", "/drivers/123/location", `{"lat":1,"lng":2,"accuracy_m":10}`, http.StatusBadRequest},
		{"missing_fields", "/drivers/11111111-1111-1111-1111-111111111111/location", `{"lat":1}`, http.StatusBadRequest},
		{"ok", "/drivers/11111111-1111-1111-1111-111111111111/location", `{"lat":1,"lng":2,"accuracy_m":10}`, http.StatusOK},
	}

	r := setupDriverRouter()
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
