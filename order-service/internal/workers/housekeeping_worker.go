package workers

import (
	"context"
	"log"
	"time"

	"github.com/FC4RICA/hong-commerce/order-service/internal/services"
)

type HousekeepingWorker struct {
	svc      services.OrderService
	interval time.Duration
}

func NewHousekeepingWorker(svc services.OrderService, interval time.Duration) *HousekeepingWorker {
	return &HousekeepingWorker{
		svc:      svc,
		interval: interval,
	}
}

func (w *HousekeepingWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Printf("Housekeeping worker started with interval %v", w.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Housekeeping worker stopping...")
			return
		case <-ticker.C:
			if err := w.svc.ProcessTimeoutOrders(ctx); err != nil {
				log.Printf("housekeeping error: %v", err)
			}
		}
	}
}
