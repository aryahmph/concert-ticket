//go:build !loadtest

package order

import (
	"context"
	"github.com/aryahmph/concert-ticket/pkg"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/spf13/cast"
)

func createVirtualAccountPayment(ctx context.Context, id string, amount uint32) (string, error) {
	ctx = pkg.TraceSpanStart(ctx, "payment.createVirtualAccountPayment")
	defer pkg.TraceSpanFinish(ctx)

	req := &coreapi.ChargeReq{
		PaymentType: coreapi.PaymentTypeBankTransfer,
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  id,
			GrossAmt: cast.ToInt64(amount),
		},
		BankTransfer: &coreapi.BankTransferDetails{
			Bank: midtrans.BankBri,
		},
	}

	res, err := coreapi.ChargeTransaction(req)
	if err != nil {
		return "", err.RawError
	}

	return res.VaNumbers[0].VANumber, nil
}
