package order

import (
	"context"
	"encoding/json"
	"github.com/aryahmph/concert-ticket/feature/shared"
	"github.com/aryahmph/concert-ticket/pkg"
	"github.com/hibiken/asynq"
	"log/slog"
)

func CancellationHandler(ctx context.Context, t *asynq.Task) error {
	var (
		lvState1       = shared.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		lvState2       = shared.LogEventStateUpdateDB
		lfState2Status = "state_2_update_order_status"

		lf = []slog.Attr{pkg.LogEventName("CancelOrder")}
	)

	ctx = pkg.TraceSpanStart(ctx, "task.CancellationHandler")
	defer pkg.TraceSpanFinish(ctx)

	/*------------------------------------
	| Step 1 : Decode request
	* ----------------------------------*/
	lf = append(lf, pkg.LogEventState(lvState1))

	var payload cancellationTask
	err := json.Unmarshal(t.Payload(), &payload)
	if err != nil {
		pkg.TraceSpanError(ctx, err)
		lf = append(lf, pkg.LogStatusFailed(lfState1Status))
		pkg.LogErrorWithContext(ctx, err, lf)
		return err
	}

	lf = append(lf,
		pkg.LogStatusSuccess(lfState1Status),
		pkg.LogEventPayload(payload),
	)

	/*------------------------------------
	| Step 2 : Update order
	* ----------------------------------*/
	lf = append(lf, pkg.LogEventState(lvState2))

	err = updateOrderStatusToCancel(ctx, payload.ID)
	if err != nil {
		if err == errOrderNotFound || err == errTicketNotFound {
			return nil
		}

		pkg.TraceSpanError(ctx, err)
		lf = append(lf, pkg.LogStatusFailed(lfState2Status))
		pkg.LogErrorWithContext(ctx, err, lf)
		return err
	}

	lf = append(lf, pkg.LogStatusSuccess(lfState2Status))

	pkg.LogInfoWithContext(ctx, "success cancel order", lf)
	return nil
}
