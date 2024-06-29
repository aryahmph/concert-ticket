up-local:
	docker compose -f deploy/compose/local.yaml up -d

down-local:
	docker compose -f deploy/compose/local.yaml down

restart-local:
	docker compose -f deploy/compose/local.yaml restart

http:
	./concert-ticket serve-http

worker:
	./concert-ticket serve-worker