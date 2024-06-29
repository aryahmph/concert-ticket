package order

import "time"

type orderEntity struct {
	ID         string
	CategoryID uint8
	Email      string
	VaCode     string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
