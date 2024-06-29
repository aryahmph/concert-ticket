package order

import "errors"

var (
	errTicketNotFound = errors.New("ticket: not found")
	errOrderExist     = errors.New("order: already exist")
	errOrderNotFound  = errors.New("order: not found")
)
