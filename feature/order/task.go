package order

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aryahmph/concert-ticket/pkg"
	"github.com/hibiken/asynq"
	"time"
)

const (
	CancellationTaskName    = "cancellation"
	cancellationTaskId      = "cancellation:%s"
	cancellationTaskTimeout = 10 * time.Second
)

func newOrderCancellationTask(ctx context.Context, id string) error {
	ctx = pkg.TraceSpanStart(ctx, "task.newOrderCancellationTask")
	defer pkg.TraceSpanFinish(ctx)

	marshal, err := json.Marshal(cancellationTask{ID: id})
	if err != nil {
		pkg.TraceSpanError(ctx, err)
		return err
	}

	_, err = queueClient.EnqueueContext(ctx, asynq.NewTask(
		CancellationTaskName,
		marshal,
		asynq.ProcessAt(time.Now().Add(orderCancellationDuration)),
		asynq.TaskID(fmt.Sprintf(cancellationTaskId, id)),
		asynq.MaxRetry(2),
		asynq.Timeout(cancellationTaskTimeout),
	))
	if err != nil {
		pkg.TraceSpanError(ctx, err)
	}

	return err
}
