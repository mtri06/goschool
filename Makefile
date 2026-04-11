run:
	go run cmd/api/main.go

dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml --env-file .env.dev up --build --attach app

prod:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml --env-file .env.prod up --build -d

test/unit:
	go test ./...

test/integration:
	docker compose -f docker-compose.yml -f docker-compose.test.yml --env-file .env.test up --build -d --wait
	go test -tags integration -timeout 120s ./integration/... ; \
	docker compose -f docker-compose.yml -f docker-compose.test.yml --env-file .env.test down

down:
	docker compose down

down/volumes:
	docker compose down -v
