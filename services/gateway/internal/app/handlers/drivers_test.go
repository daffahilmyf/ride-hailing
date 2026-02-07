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
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
)

type captureMatchingClient struct {
	lastStatus *matchingv1.UpdateDriverStatusRequest
}

type captureLocationClient struct {
	lastLocation *locationv1.UpdateDriverLocationRequest
	lastNearby   *locationv1.ListNearbyDriversRequest
}

func (f *captureMatchingClient) UpdateDriverStatus(ctx context.Context, in *matchingv1.UpdateDriverStatusRequest, opts ...grpc.CallOption) (*matchingv1.UpdateDriverStatusResponse, error) {
	f.lastStatus = in
	return &matchingv1.UpdateDriverStatusResponse{Status: "OK"}, nil
}

func (f *captureLocationClient) UpdateDriverLocation(ctx context.Context, in *locationv1.UpdateDriverLocationRequest, opts ...grpc.CallOption) (*locationv1.UpdateDriverLocationResponse, error) {
	f.lastLocation = in
	return &locationv1.UpdateDriverLocationResponse{Status: "OK"}, nil
}

func (f *captureLocationClient) ListNearbyDrivers(ctx context.Context, in *locationv1.ListNearbyDriversRequest, opts ...grpc.CallOption) (*locationv1.ListNearbyDriversResponse, error) {
	f.lastNearby = in
	return &locationv1.ListNearbyDriversResponse{}, nil
}

func setupDriverRouter(matching *captureMatchingClient, location *captureLocationClient) *gin.Engine {
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
	r.POST("/drivers/:driver_id/status", UpdateDriverStatus(matching))
	r.POST("/drivers/:driver_id/location", UpdateDriverLocation(location, ""))
	r.POST("/drivers/nearby", ListNearbyDrivers(location, ""))
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

	matching := &captureMatchingClient{}
	location := &captureLocationClient{}
	r := setupDriverRouter(matching, location)
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

	matching := &captureMatchingClient{}
	location := &captureLocationClient{}
	r := setupDriverRouter(matching, location)
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

func TestListNearbyDriversValidation(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		status int
	}{
		{"missing_fields", `{"lat":1}`, http.StatusBadRequest},
		{"ok", `{"lat":1,"lng":2,"radius_m":1000}`, http.StatusOK},
	}

	matching := &captureMatchingClient{}
	location := &captureLocationClient{}
	r := setupDriverRouter(matching, location)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/drivers/nearby", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, w.Code)
			}
		})
	}
}

func TestUpdateDriverLocationMetadataHeaders(t *testing.T) {
	location := &captureLocationClient{}
	r := setupDriverRouter(&captureMatchingClient{}, location)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/drivers/11111111-1111-1111-1111-111111111111/location", bytes.NewBufferString(`{"lat":1,"lng":2,"accuracy_m":10}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Trace-Id", "trace-2")
	req.Header.Set("X-Request-Id", "req-2")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
	if location.lastLocation == nil {
		t.Fatalf("expected location request")
	}
	if location.lastLocation.GetTraceId() != "trace-2" {
		t.Fatalf("expected trace id")
	}
	if location.lastLocation.GetRequestId() != "req-2" {
		t.Fatalf("expected request id")
	}
}
