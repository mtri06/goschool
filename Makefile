run:
	go run cmd/api/main.go

dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml --env-file .env.dev up --build -d

dev/logs:
	docker logs -f goschool-app-dev

dev/stop:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml --env-file .env.dev stop

unit_test:
	go test ./...

test:
	docker compose -f docker-compose.yml -f docker-compose.test.yml --env-file .env.test up --build -d --wait
	go test -tags integration -timeout 120s ./integration/... ; \
	docker compose -f docker-compose.yml -f docker-compose.test.yml --env-file .env.test down

test/stop:
	docker compose -f docker-compose.yml -f docker-compose.test.yml --env-file .env.test stop

prod:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml --env-file .env.prod up --build -d

dev/down_volumes:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml --env-file .env.dev down -v
