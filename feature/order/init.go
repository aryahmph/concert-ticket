package order

import (
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var (
	db             *pgxpool.Pool
	cache          *redis.Client
	queueClient    *asynq.Client
	queueInspector *asynq.Inspector
)

func SetDB(pool *pgxpool.Pool) {
	if pool == nil {
		panic("cannot assign nil db pool")
	}

	db = pool
}

func SetCache(pool *redis.Client) {
	if pool == nil {
		panic("cannot assign nil cache pool")
	}

	cache = pool
}

func SetQueueClient(client *asynq.Client) {
	if client == nil {
		panic("cannot assign nil queue client")
	}

	queueClient = client
}

func SetQueueInspector(inspector *asynq.Inspector) {
	if inspector == nil {
		panic("cannot assign nil queue inspector")
	}

	queueInspector = inspector
}
