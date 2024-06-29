package pkg

import (
	"context"
	"github.com/aryahmph/concert-ticket/feature/shared"
	"github.com/exaring/otelpgx"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func InitDatabase(ctx context.Context, cfg shared.Config) *pgxpool.Pool {
	dbCfg, err := pgxpool.ParseConfig(cfg.Database.ConnStr())
	if err != nil {
		panic(err)
	}

	dbCfg.ConnConfig.Tracer = otelpgx.NewTracer()

	db, err := pgxpool.NewWithConfig(ctx, dbCfg)
	if err != nil {
		panic(err)
	}

	return db
}

func InitCache(cfg shared.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Cache.Addr,
		MinIdleConns: int(cfg.Cache.MinConn),
	})

	return client
}

func InitClientQueue(cfg shared.Config) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Cache.Addr,
		PoolSize: int(cfg.Cache.MinConn),
	})
}

func InitInspectorQueue(cfg shared.Config) *asynq.Inspector {
	return asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     cfg.Cache.Addr,
		PoolSize: int(cfg.Cache.MinConn),
	})
}
