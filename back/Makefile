.PHONY: fmt lint test build check up down clean

fmt:
	docker compose run --rm fmt

lint:
	docker compose run --rm lint

test:
	docker compose run --rm test

build:
	docker compose build app

check:
	$(MAKE) fmt
	$(MAKE) lint
	$(MAKE) test
	$(MAKE) build

up:
	docker compose up --build

down:
	docker compose down

clean:
	docker compose down
	docker compose rm -f