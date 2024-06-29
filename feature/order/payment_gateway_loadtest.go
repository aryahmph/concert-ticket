//go:build loadtest

package order

import (
	"context"
	"github.com/aryahmph/concert-ticket/pkg"
	"time"
)

func createVirtualAccountPayment(ctx context.Context, id string, amount uint32) (string, error) {
	ctx = pkg.TraceSpanStart(ctx, "payment.createVirtualAccountPayment")
	defer pkg.TraceSpanFinish(ctx)

	time.Sleep(100 * time.Millisecond)
	return "GENERATED-VA", nil
}
