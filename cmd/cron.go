package cmd

import (
	"context"
	"github.com/aryahmph/concert-ticket/feature/category"
	"github.com/aryahmph/concert-ticket/feature/shared"
	"github.com/aryahmph/concert-ticket/pkg"
	"log"
	"time"
)

func runCronCmd(ctx context.Context) {
	cfg := shared.LoadConfig("config/cron.yaml")

	// Dependency
	db := pkg.InitDatabase(ctx, cfg)
	defer db.Close()

	cache := pkg.InitCache(cfg)
	defer cache.Close()

	category.SetDB(db)
	category.SetCache(cache)

	ticker := time.NewTicker(3 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				category.UpdateListCategories(ctx)
			}
		}
	}()

	log.Println("cron started")

	<-ctx.Done()
	ticker.Stop()

	log.Println("cron shutdown")
}
