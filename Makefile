run:
	go run cmd/api/main.go

dev:
	docker compose -f docker-compose.dev.yml --env-file .env.dev up --build -d
	$(MAKE) dev/logs 

dev/logs:
	docker logs -f goschool-app-dev

dev/stop:
	docker compose -f docker-compose.dev.yml stop

dev/volume_down:
	docker compose -f docker-compose.dev.yml down -v

test/unit:
	docker compose -f docker-compose.test.yml --env-file .env.test run --build --rm app_unit_test

test/integration:
	docker compose -f docker-compose.test.yml --env-file .env.test run --build --rm app_integration_test

prod:
	docker compose -f docker-compose.prod.yml --env-file .env.prod up --build -d
