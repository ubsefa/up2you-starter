.PHONY: setup up down reset logs ps smoke package config validate new

setup:
	@if [ ! -f .env ]; then cp .env.example .env; fi
	docker compose --env-file .env config --quiet

up:
	docker compose up -d

down:
	docker compose down

reset:
	docker compose down -v
	docker compose up -d

logs:
	docker compose logs -f

ps:
	docker compose ps

smoke:
	./scripts/smoke-test.sh

package:
	./scripts/package-app.sh examples/my-todo

config:
	docker compose config --quiet

validate:
	./scripts/validate-examples.sh

new:
	@test -n "$(APP)" || { echo "usage: make new APP=<app-key>"; exit 1; }
	./scripts/new-app.sh "$(APP)"
