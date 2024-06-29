package order

import (
	"encoding/json"
	"fmt"
	"github.com/aryahmph/concert-ticket/feature/category"
	"github.com/aryahmph/concert-ticket/feature/shared"
	"github.com/aryahmph/concert-ticket/pkg"
	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"
	"log/slog"
	"net/http"
)

func HttpRoute(mux *http.ServeMux) {
	mux.HandleFunc("POST /orders", createOrderHandler)
	mux.HandleFunc("POST /payments/callback", paymentCallbackHandler)
}

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	var (
		lvState1       = shared.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		lvState2       = shared.LogEventStateFetchDB
		lfState2Status = "state_2_check_user_has_order_status"

		lvState3       = shared.LogEventStateSetCache
		lfState3Status = "state_3_lock_email_status"

		lvState4       = shared.LogEventStateInsertDB
		lfState4Status = "state_4_insert_order_status"

		lvState5       = shared.LogEventStateCreatePayment
		lfState5Status = "state_5_create_payment_status"

		lvState6       = shared.LogEventStateUpdateDB
		lfState6Status = "state_6_update_order_status"

		lvState7       = shared.LogEventStateSetCache
		lfState7Status = "state_7_insert_task_status"

		ctx = pkg.TraceSpanStart(r.Context(), "http.createOrderHandler")

		lf = []slog.Attr{
			pkg.LogEventName("CreateOrder"),
		}
	)

	defer pkg.TraceSpanFinish(ctx)

	/*------------------------------------
	| Step 1 : Decode request
	* ----------------------------------*/
	lf = append(lf, pkg.LogEventState(lvState1))

	var req createOrderRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		pkg.TraceSpanError(ctx, err)
		lf = append(lf, pkg.LogStatusFailed(lfState1Status))
		pkg.LogWarnWithContext(ctx, "invalid request", err, lf)
		shared.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	lf = append(lf,
		pkg.LogStatusSuccess(lfState1Status),
		pkg.LogEventPayload(req),
	)

	/*------------------------------------
	| Step 2 : Check user order
	* ----------------------------------*/
	lf = append(lf, pkg.LogEventState(lvState2))

	isUserAlreadyOrder, err := isUserHasOrder(ctx, req.Email)
	if err != nil {
		pkg.TraceSpanError(ctx, err)
		lf = append(lf, pkg.LogStatusFailed(lfState2Status))
		pkg.LogErrorWithContext(ctx, err, lf)
		shared.WriteInternalServerErrorResponse(w)
		return
	}

	if isUserAlreadyOrder {
		shared.WriteErrorResponse(w, http.StatusConflict, errOrderExist)
		return
	}

	lf = append(lf, pkg.LogStatusSuccess(lfState2Status))

	/*------------------------------------
	| Step 3 : Lock email
	* ----------------------------------*/
	lf = append(lf, pkg.LogEventState(lvState3))

	emailLock, err := cache.SetNX(ctx, fmt.Sprintf(orderLockCacheKey, req.Email), true, orderLockDuration).Result()
	if err != nil {
		pkg.TraceSpanError(ctx, err)
		lf = append(lf, pkg.LogStatusFailed(lfState3Status))
		pkg.LogErrorWithContext(ctx, err, lf)
		shared.WriteInternalServerErrorResponse(w)
		return
	}

	if !emailLock {
		shared.WriteErrorResponse(w, http.StatusConflict, errOrderExist)
		return
	}

	lf = append(lf, pkg.LogStatusSuccess(lfState3Status))

	/*------------------------------------
	| Step 4 : Insert order
	* ----------------------------------*/
	lf = append(lf, pkg.LogEventState(lvState4))

	tx, err := db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadUncommitted,
	})
	if err != nil {
		pkg.TraceSpanError(ctx, err)
		lf = append(lf, pkg.LogStatusFailed(lfState4Status))
		pkg.LogErrorWithContext(ctx, err, lf)
		shared.WriteInternalServerErrorResponse(w)
		return
	}

	order := orderEntity{
		ID:         ulid.Make().String(),
		CategoryID: req.CategoryID,
		Email:      req.Email,
	}

	err = insertOrder(ctx, tx, order)
	if err != nil {
		_ = tx.Rollback(ctx)

		if err == errTicketNotFound {
			shared.WriteErrorResponse(w, http.StatusNotFound, err)
			return
		}

		pkg.TraceSpanError(ctx, err)
		lf = append(lf, pkg.LogStatusFailed(lfState4Status))
		pkg.LogErrorWithContext(ctx, err, lf)
		shared.WriteInternalServerErrorResponse(w)
		return
	}

	lf = append(lf, pkg.LogStatusSuccess(lfState4Status))

	/*------------------------------------
	| Step 5 : Create payment
	* ----------------------------------*/
	lf = append(lf, pkg.LogEventState(lvState5))
	price := category.Categories[req.CategoryID].Price

	vaCode, err := createVirtualAccountPayment(ctx, order.ID, price)
	if err != nil {
		_ = tx.Rollback(ctx)
		pkg.TraceSpanError(ctx, err)
		lf = append(lf, pkg.LogStatusFailed(lfState5Status))
		pkg.LogErrorWithContext(ctx, err, lf)
		shared.WriteInternalServerErrorResponse(w)
		return
	}

	lf = append(lf, pkg.LogStatusSuccess(lfState5Status))

	/*------------------------------------
	| Step 6 : Update order
	* ----------------------------------*/
	lf = append(lf, pkg.LogEventState(lvState6))

	err = updateOrderVaCode(ctx, tx, order.ID, vaCode)
	if err != nil {
		_ = tx.Rollback(ctx)

		if err == errOrderNotFound {
			shared.WriteErrorResponse(w, http.StatusNotFound, err)
			return
		}

		pkg.TraceSpanError(ctx, err)
		lf = append(lf, pkg.LogStatusFailed(lfState6Status))
		pkg.LogErrorWithContext(ctx, err, lf)
		shared.WriteInternalServerErrorResponse(w)
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		pkg.TraceSpanError(ctx, err)
		_ = tx.Rollback(ctx)
		lf = append(lf, pkg.LogStatusFailed(lfState6Status))
		pkg.LogErrorWithContext(ctx, err, lf)
		shared.WriteInternalServerErrorResponse(w)
		return
	}

	lf = append(lf, pkg.LogStatusSuccess(lfState6Status))

	/*------------------------------------
	| Step 7 : Insert task
	* ----------------------------------*/
	lf = append(lf, pkg.LogEventState(lvState7))

	err = newOrderCancellationTask(ctx, order.ID)
	if err != nil {
		pkg.TraceSpanError(ctx, err)
		lf = append(lf, pkg.LogStatusFailed(lfState7Status))
		pkg.LogErrorWithContext(ctx, err, lf)
		shared.WriteInternalServerErrorResponse(w)
		return
	}

	lf = append(lf, pkg.LogStatusSuccess(lfState7Status))

	pkg.LogInfoWithContext(ctx, "success create order", lf)
	shared.WriteSuccessResponse(w, http.StatusCreated,
		createOrderResponse{
			ID:     order.ID,
			Total:  price,
			VaCode: vaCode,
		})
}

func paymentCallbackHandler(w http.ResponseWriter, r *http.Request) {
	var (
		lvState1       = shared.LogEventStateDecodeRequest
		lfState1Status = "state_1_decode_request_status"

		lvState2       = shared.LogEventStateUpdateDB
		lfState2Status = "state_2_update_order_status"

		lvState3       = shared.LogEventStateSetCache
		lfState3Status = "state_3_delete_task_status"

		ctx = pkg.TraceSpanStart(r.Context(), "http.paymentCallbackHandler")

		lf = []slog.Attr{
			pkg.LogEventName("PaymentCallback"),
		}
	)

	defer pkg.TraceSpanFinish(ctx)

	/*------------------------------------
	| Step 1 : Decode request
	* ----------------------------------*/
	lf = append(lf, pkg.LogEventState(lvState1))

	var req paymentNotificationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		pkg.TraceSpanError(ctx, err)
		lf = append(lf, pkg.LogStatusFailed(lfState1Status))
		pkg.LogWarnWithContext(ctx, "invalid request", err, lf)
		shared.WriteErrorResponse(w, http.StatusBadRequest, err)
	}

	lf = append(lf,
		pkg.LogStatusSuccess(lfState1Status),
		pkg.LogEventPayload(req),
	)

	/*------------------------------------
	| Step 2: Update order
	* ----------------------------------*/
	lf = append(lf, pkg.LogEventState(lvState2))

	err = updateOrderStatusToComplete(ctx, req.OrderID)
	if err != nil {
		if err == errOrderNotFound || err == errTicketNotFound {
			shared.WriteSuccessResponse(w, http.StatusOK, nil)
			return
		}

		pkg.TraceSpanError(ctx, err)
		lf = append(lf, pkg.LogStatusFailed(lfState2Status))
		pkg.LogErrorWithContext(ctx, err, lf)
		shared.WriteInternalServerErrorResponse(w)
		return
	}

	lf = append(lf, pkg.LogStatusSuccess(lfState2Status))

	/*------------------------------------
	| Step 3: Delete task
	* ----------------------------------*/
	lf = append(lf, pkg.LogEventState(lvState3))

	err = queueInspector.DeleteTask("default", fmt.Sprintf(cancellationTaskId, req.OrderID))
	if err != nil {
		pkg.TraceSpanError(ctx, err)
		lf = append(lf, pkg.LogStatusFailed(lfState3Status))
		pkg.LogErrorWithContext(ctx, err, lf)
		shared.WriteInternalServerErrorResponse(w)
		return
	}

	lf = append(lf, pkg.LogStatusSuccess(lfState3Status))

	pkg.LogInfoWithContext(ctx, "success handle payment callback", lf)
	shared.WriteSuccessResponse(w, http.StatusOK, nil)
}
