include .env
export

export PROJECT_ROOT=${shell pwd}

env-up:
	docker compose up -d todolist-postgres

env-down:
	docker compose down todolist-postgres

env-cleanup:
	@read -p "Отчистить данные Postgres??? [y/n]: " ans; \
	if [ "$$ans" = "y" ]; then \
	  docker compose down todolist-postgres port-forwarder && \
	  sudo rm -rf ${PROJECT_ROOT}/_out/pgdata && \
	  echo "Очистилось удачно!"; \
	else \
	  echo "Очистилось неудачно!"; \
	fi

migration-create:
	@if [ -z "$(seq)" ]; then \
		echo "Не задан параметр - name migrate."; \
		exit 1; \
	fi; \
	docker compose run --rm todolist-postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migration-action:
	@if [ -z "$(action)" ]; then \
		echo "Не задан параметр - up/down (1,2,...)."; \
		exit 1; \
	fi; \
	docker compose run --rm todolist-postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@todolist-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"

env-port-forward:
	@docker compose up -d port-forwarder

env-port-close:
	@docker compose down port-forwarder

logs-cleanup:
	@read -p "Отчистить данные Logger'а??? [y/n]: " ans; \
	if [ "$$ans" = "y" ]; then \
	  sudo rm -rf ${PROJECT_ROOT}/_out/logs && \
	  echo "Очистилось удачно!"; \
	else \
	  echo "Очистилось неудачно!"; \
	fi

todolist-run:
	@export LOGGER_FOLDER=${PROJECT_ROOT}/_out/logs && \
	export POSTGRES_HOST=localhost && \
	go mod tidy && \
	go run ${PROJECT_ROOT}/cmd/todolist/main.go
