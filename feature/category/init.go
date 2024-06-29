package category

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var (
	db    *pgxpool.Pool
	cache *redis.Client
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
