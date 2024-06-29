package order

import (
	"context"
	"github.com/aryahmph/concert-ticket/pkg"
	"github.com/jackc/pgx/v5"
)

func isUserHasOrder(ctx context.Context, email string) (isExist bool, err error) {
	ctx = pkg.TraceSpanStart(ctx, "repo.isUserHasOrder")
	defer pkg.TraceSpanFinish(ctx)

	q := `SELECT EXISTS(SELECT 1 FROM orders WHERE email = $1 AND status IN ('created', 'completed'));`

	err = db.QueryRow(ctx, q, email).Scan(&isExist)
	if err != nil {
		pkg.TraceSpanError(ctx, err)
		return
	}

	return
}

func insertOrder(ctx context.Context, tx pgx.Tx, order orderEntity) (err error) {
	ctx = pkg.TraceSpanStart(ctx, "repo.insertOrder")
	defer pkg.TraceSpanFinish(ctx)

	q1 := `
		UPDATE tickets 
		SET order_id = $2 
		WHERE id IN (
			SELECT id 
			FROM tickets 
			WHERE category_id = $1 AND order_id IS NULL 
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		);
	`
	q2 := `INSERT INTO orders (id, category_id, email) VALUES ($1, $2, $3);`

	cmd, err := tx.Exec(ctx, q1, order.CategoryID, order.ID)
	if err != nil {
		pkg.TraceSpanError(ctx, err)
		return
	}

	if cmd.RowsAffected() == 0 {
		err = errTicketNotFound
		return
	}

	_, err = tx.Exec(ctx, q2, order.ID, order.CategoryID, order.Email)
	if err != nil {
		pkg.TraceSpanError(ctx, err)
		return
	}

	return
}

func updateOrderVaCode(ctx context.Context, tx pgx.Tx, id string, vaCode string) (err error) {
	ctx = pkg.TraceSpanStart(ctx, "repo.updateOrderVaCode")
	defer pkg.TraceSpanFinish(ctx)

	q := `UPDATE orders SET va_code = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2;`

	cmd, err := tx.Exec(ctx, q, vaCode, id)
	if err != nil {
		pkg.TraceSpanError(ctx, err)
		return
	}

	if cmd.RowsAffected() == 0 {
		err = errOrderNotFound
	}

	return
}

func updateOrderStatusToComplete(ctx context.Context, id string) (err error) {
	ctx = pkg.TraceSpanStart(ctx, "repo.updateOrderStatusToComplete")
	defer pkg.TraceSpanFinish(ctx)

	q := `UPDATE orders SET status = 'completed', updated_at = CURRENT_TIMESTAMP WHERE id = $1 AND status = 'created';`

	cmd, err := db.Exec(ctx, q, id)
	if err != nil {
		pkg.TraceSpanError(ctx, err)
		return
	}

	if cmd.RowsAffected() == 0 {
		err = errOrderNotFound
	}

	return
}

func updateOrderStatusToCancel(ctx context.Context, id string) (err error) {
	ctx = pkg.TraceSpanStart(ctx, "repo.updateOrderStatusToCancel")
	defer pkg.TraceSpanFinish(ctx)

	q1 := `UPDATE orders SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP WHERE id = $1 AND status = 'created';`
	q2 := `UPDATE tickets SET order_id = NULL WHERE order_id = $1;`

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	cmd, err := tx.Exec(ctx, q1, id)
	if err != nil {
		_ = tx.Rollback(ctx)
		pkg.TraceSpanError(ctx, err)
		return
	}

	if cmd.RowsAffected() == 0 {
		err = errOrderNotFound
		_ = tx.Rollback(ctx)
		return
	}

	cmd, err = tx.Exec(ctx, q2, id)
	if err != nil {
		_ = tx.Rollback(ctx)
		pkg.TraceSpanError(ctx, err)
		return
	}

	if cmd.RowsAffected() == 0 {
		err = errTicketNotFound
		_ = tx.Rollback(ctx)
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		_ = tx.Rollback(ctx)
		pkg.TraceSpanError(ctx, err)
	}

	return
}
