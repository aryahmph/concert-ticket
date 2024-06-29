package order

import "time"

const (
	orderLockCacheKey = "order_lock:%s"
	orderLockDuration = 2 * 60 * time.Second
)

var (
	orderCancellationDuration = 15 * time.Second
)

func SetOrderCancellationDuration(d time.Duration) {
	if d > 0 {
		orderCancellationDuration = d
	}
}
