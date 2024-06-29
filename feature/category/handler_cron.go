package category

import (
	"context"
	"encoding/json"
	"github.com/aryahmph/concert-ticket/feature/shared"
	"github.com/aryahmph/concert-ticket/pkg"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

func UpdateListCategories(ctx context.Context) {
	var (
		lvState1       = shared.LogEventStateFetchDB
		lfState1Status = "state_1_count_ticket_status"

		lvState2       = shared.LogEventStateSetCache
		lfState2Status = "state_2_set_cache_status"

		lf = []slog.Attr{
			pkg.LogEventName("UpdateListCategories"),
		}
	)

	/*------------------------------------
	| Step 1 : Count ticket
	* ----------------------------------*/
	lf = append(lf, pkg.LogEventState(lvState1))

	categoryIds, total, err := countAvailableTicketByCategoryGroup(ctx)
	if err != nil {
		lf = append(lf, pkg.LogStatusFailed(lfState1Status))
		pkg.LogErrorWithContext(ctx, err, lf)
		return
	}

	lf = append(lf, pkg.LogStatusSuccess(lfState1Status))

	/*------------------------------------
	| Step 2 : Set cache
	* ----------------------------------*/
	lf = append(lf, pkg.LogEventState(lvState2))

	response := make([]listCategoriesResponse, len(categoryIds))
	for i, categoryId := range categoryIds {
		response[i] = listCategoriesResponse{
			ID:    categoryId,
			Total: total[i],
		}
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		lf = append(lf, pkg.LogStatusFailed(lfState2Status))
		pkg.LogErrorWithContext(ctx, err, lf)
		return
	}

	err = cache.Set(ctx, listCategoriesCacheKey, bytes, redis.KeepTTL).Err()
	if err != nil {
		lf = append(lf, pkg.LogStatusFailed(lfState2Status))
		pkg.LogErrorWithContext(ctx, err, lf)
		return
	}
}
