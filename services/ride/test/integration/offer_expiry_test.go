//go:build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/adapters/db"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/app/usecase"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/app/workers"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/domain"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/ports/outbound"
	"go.uber.org/zap"
)

func TestOfferExpiryWorker(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set")
	}

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	repo := db.NewRideOfferRepo(conn)
	offer := domain.NewRideOffer("ride-1", "driver-1", -1*time.Second)
	require.NoError(t, repo.Create(context.Background(), toOutboundOffer(offer)))

	logger, _ := zap.NewDevelopment()
	uc := &usecase.RideService{Offers: repo, OfferMetrics: &usecase.OfferMetrics{}}
	worker := &workers.OfferExpiryWorker{
		Repo:     repo,
		Usecase:  uc,
		Logger:   logger,
		Interval: 50 * time.Millisecond,
		Batch:    10,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.Run(ctx)

	time.Sleep(200 * time.Millisecond)
	updated, err := repo.Get(context.Background(), offer.ID)
	require.NoError(t, err)
	require.Equal(t, string(domain.OfferExpired), updated.Status)
}

func toOutboundOffer(offer domain.RideOffer) outbound.RideOffer {
	return outbound.RideOffer{
		ID:        offer.ID,
		RideID:    offer.RideID,
		DriverID:  offer.DriverID,
		Status:    string(offer.Status),
		ExpiresAt: offer.ExpiresAt.Unix(),
		CreatedAt: offer.CreatedAt.Unix(),
	}
}
