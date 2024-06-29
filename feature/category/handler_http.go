package category

import (
	"encoding/json"
	"github.com/aryahmph/concert-ticket/feature/shared"
	"github.com/aryahmph/concert-ticket/pkg"
	"log/slog"
	"net/http"
)

func HttpRoute(mux *http.ServeMux) {
	mux.HandleFunc("GET /categories", listCategoriesHandler)
}

func listCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	var (
		lvState1       = shared.LogEventStateGetCache
		lfState1Status = "state_1_fetch_cache_status"

		ctx = r.Context()

		lf = []slog.Attr{
			pkg.LogEventName("ListCategories"),
		}
	)

	/*------------------------------------
	| Step 1 : Fetch cache
	* ----------------------------------*/
	lf = append(lf, pkg.LogEventState(lvState1))

	dataCache, err := cache.Get(ctx, listCategoriesCacheKey).Result()
	if err != nil {
		lf = append(lf, pkg.LogStatusFailed(lfState1Status))
		pkg.LogErrorWithContext(ctx, err, lf)
		shared.WriteInternalServerErrorResponse(w)
	}

	var ticketCache []listCategoriesResponse
	err = json.Unmarshal([]byte(dataCache), &ticketCache)
	if err != nil {
		lf = append(lf, pkg.LogStatusFailed(lfState1Status))
		pkg.LogErrorWithContext(ctx, err, lf)
		shared.WriteInternalServerErrorResponse(w)
	}

	response := []listCategoriesResponse{
		{ID: 1, Name: "VIP", Price: 3_800_000},
		{ID: 2, Name: "PLATINUM", Price: 3_400_000},
		{ID: 3, Name: "CAT 1", Price: 2_900_000},
		{ID: 4, Name: "CAT 2", Price: 2_600_000},
		{ID: 5, Name: "CAT 3", Price: 2_100_000},
	}

	for _, ticket := range ticketCache {
		response[ticket.ID-1].Total = ticket.Total
	}

	shared.WriteSuccessResponse(w, http.StatusOK, response)
}
