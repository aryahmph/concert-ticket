package category

import (
	"context"
)

func countAvailableTicketByCategoryGroup(ctx context.Context) (categoryIds []uint8, total []uint16, err error) {
	q := `SELECT category_id, COUNT(*) FROM tickets WHERE order_id IS NULL GROUP BY category_id;`

	rows, err := db.Query(ctx, q)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var categoryId uint8
		var count uint16

		err = rows.Scan(&categoryId, &count)
		if err != nil {
			return
		}

		categoryIds = append(categoryIds, categoryId)
		total = append(total, count)
	}

	return
}
