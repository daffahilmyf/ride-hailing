package handlers

import (
	"context"
	"testing"
	"time"

	locationv1 "github.com/daffahilmyf/ride-hailing/proto/location/v1"
	"github.com/daffahilmyf/ride-hailing/services/location/internal/app/usecase"
	"github.com/daffahilmyf/ride-hailing/services/location/internal/ports/outbound"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeRepo struct {
	location outbound.Location
	err      error
}

func (f *fakeRepo) Upsert(_ context.Context, _ outbound.Location, _ time.Duration) error {
	return nil
}

func (f *fakeRepo) Get(_ context.Context, _ string) (outbound.Location, error) {
	if f.err != nil {
		return outbound.Location{}, f.err
	}
	return f.location, nil
}

func TestGetDriverLocation(t *testing.T) {
	tests := []struct {
		name       string
		repoErr    error
		wantCode   codes.Code
		wantDriver string
	}{
		{"not_found", outbound.ErrNotFound, codes.NotFound, ""},
		{"ok", nil, codes.OK, "driver-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{
				err: tt.repoErr,
				location: outbound.Location{
					DriverID:   "driver-1",
					Lat:        1,
					Lng:        2,
					AccuracyM:  3,
					RecordedAt: time.Unix(10, 0).UTC(),
				},
			}
			svc := &usecase.LocationService{Repo: repo}
			server := &LocationServer{usecase: svc}

			resp, err := server.GetDriverLocation(context.Background(), &locationv1.GetDriverLocationRequest{
				DriverId: "driver-1",
			})
			if tt.wantCode == codes.OK {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if resp.GetDriverId() != tt.wantDriver {
					t.Fatalf("unexpected driver_id: %s", resp.GetDriverId())
				}
				return
			}
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if status.Code(err) != tt.wantCode {
				t.Fatalf("unexpected code: %v", status.Code(err))
			}
		})
	}
}
