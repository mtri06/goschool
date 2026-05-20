run:
	go run cmd/api/main.go

dev:
	docker compose -f docker-compose.dev.yml --env-file .env.dev up --build -d
	$(MAKE) dev/logs 

dev/logs:
	docker logs -f goschool-app-dev

dev/stop:
	docker compose -f docker-compose.dev.yml --env-file .env.dev stop

dev/volume_down:
	docker compose -f docker-compose.dev.yml --env-file .env.dev down -v

test/unit:
	docker compose -f docker-compose.unit-test.yml --env-file .env.test up --build --abort-on-container-exit \
	--exit-code-from unit_test --no-log-prefix --remove-orphans 

test/integration:
	docker compose -f docker-compose.test.yml --env-file .env.test up --build --abort-on-container-exit \
	--exit-code-from integration_test --attach integration_test --no-log-prefix --remove-orphans 

test/postgres:
	docker compose -f docker-compose.test.yml --env-file .env.test up postgres -d

prod:
	docker compose -f docker-compose.prod.yml --env-file .env.prod up --build -d
