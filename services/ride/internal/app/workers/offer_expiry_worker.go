package workers

import (
	"context"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/app/usecase"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/ports/outbound"
	"go.uber.org/zap"
)

type OfferExpiryWorker struct {
	Repo     outbound.RideOfferRepo
	Usecase  *usecase.RideService
	Logger   *zap.Logger
	Interval time.Duration
	Batch    int
}

func (w *OfferExpiryWorker) Run(ctx context.Context) {
	if w.Repo == nil || w.Usecase == nil || w.Logger == nil {
		return
	}
	interval := w.Interval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	batch := w.Batch
	if batch <= 0 {
		batch = 50
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.expire(ctx, batch)
		}
	}
}

func (w *OfferExpiryWorker) expire(ctx context.Context, batch int) {
	cutoff := time.Now().UTC().Unix()
	offers, err := w.Repo.ListExpired(ctx, cutoff, batch)
	if err != nil {
		w.Logger.Warn("offers.expire_scan_failed", zap.Error(err))
		return
	}
	for _, offer := range offers {
		_, err := w.Usecase.ExpireOffer(ctx, usecase.OfferActionCmd{OfferID: offer.ID})
		if err != nil {
			w.Logger.Warn("offers.expire_failed", zap.String("offer_id", offer.ID), zap.Error(err))
		}
	}
}
