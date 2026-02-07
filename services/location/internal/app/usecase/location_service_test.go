package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/location/internal/domain"
	"github.com/daffahilmyf/ride-hailing/services/location/internal/ports/outbound"
)

type fakeRepo struct {
	upsertErr   error
	getErr      error
	getLocation outbound.Location
	lastTTL     time.Duration
	lastUpsert  outbound.Location
}

func (f *fakeRepo) Upsert(_ context.Context, location outbound.Location, ttl time.Duration) error {
	f.lastUpsert = location
	f.lastTTL = ttl
	return f.upsertErr
}

func (f *fakeRepo) Get(_ context.Context, _ string) (outbound.Location, error) {
	if f.getErr != nil {
		return outbound.Location{}, f.getErr
	}
	return f.getLocation, nil
}

func (f *fakeRepo) Nearby(_ context.Context, _ float64, _ float64, _ float64, _ int) ([]outbound.NearbyDriver, error) {
	return []outbound.NearbyDriver{}, nil
}

type fakePublisher struct {
	err     error
	subject string
	payload []byte
	calls   int
}

func (f *fakePublisher) Publish(_ context.Context, subject string, payload []byte) error {
	f.calls++
	f.subject = subject
	f.payload = payload
	return f.err
}

type fakeLimiter struct {
	allowed bool
	err     error
}

func (f *fakeLimiter) Allow(_ context.Context, _ string, _ time.Duration) (bool, error) {
	return f.allowed, f.err
}

func TestUpdateDriverLocation(t *testing.T) {
	tests := []struct {
		name           string
		repoErr        error
		publishErr     error
		publishEnabled bool
		limitAllowed   bool
		limitErr       error
		minGap         time.Duration
		lat            float64
		expectErr      bool
	}{
		{"invalid_lat", nil, nil, false, true, nil, 0, 200, true},
		{"repo_error", errors.New("db"), nil, false, true, nil, 0, 1, true},
		{"publish_disabled", nil, nil, false, true, nil, 0, 1, false},
		{"publish_enabled", nil, nil, true, true, nil, 0, 1, false},
		{"rate_limited", nil, nil, false, false, nil, 10 * time.Second, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{upsertErr: tt.repoErr}
			publisher := &fakePublisher{err: tt.publishErr}
			limiter := &fakeLimiter{allowed: tt.limitAllowed, err: tt.limitErr}
			svc := &LocationService{
				Repo:           repo,
				Publisher:      publisher,
				RateLimiter:    limiter,
				PublishEnabled: tt.publishEnabled,
				LocationTTL:    10 * time.Second,
				MinUpdateGap:   tt.minGap,
			}
			_, err := svc.UpdateDriverLocation(context.Background(), "driver-1", tt.lat, 2, 1)
			if tt.expectErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.minGap > 0 && !tt.limitAllowed {
				if repo.lastUpsert.DriverID != "" {
					t.Fatal("expected no upsert when rate limited")
				}
			}
			if tt.publishEnabled && !tt.expectErr && tt.limitAllowed {
				if publisher.calls != 1 {
					t.Fatalf("expected publish call, got %d", publisher.calls)
				}
				if publisher.subject != "driver.location.updated" {
					t.Fatalf("unexpected subject: %s", publisher.subject)
				}
				var envelope domain.EventEnvelope
				if err := json.Unmarshal(publisher.payload, &envelope); err != nil {
					t.Fatalf("invalid payload: %v", err)
				}
				if envelope.Type != "driver.location.updated" {
					t.Fatalf("unexpected event type: %s", envelope.Type)
				}
			}
		})
	}
}

func TestGetDriverLocation(t *testing.T) {
	tests := []struct {
		name      string
		getErr    error
		expectErr bool
	}{
		{"not_found", outbound.ErrNotFound, true},
		{"ok", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{
				getErr: tt.getErr,
				getLocation: outbound.Location{
					DriverID:   "driver-1",
					Lat:        1,
					Lng:        2,
					AccuracyM:  3,
					RecordedAt: time.Unix(10, 0).UTC(),
				},
			}
			svc := &LocationService{Repo: repo}
			_, err := svc.GetDriverLocation(context.Background(), "driver-1")
			if tt.expectErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
