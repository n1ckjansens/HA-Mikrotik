.PHONY: dev-env dev-up dev-up-d dev-down dev-logs dev-smoke mock-online mock-offline go-test ui-test

dev-env:
	@test -f .env.dev || cp .env.dev.example .env.dev

dev-up: dev-env
	docker compose -f docker-compose.dev.yml up --build

dev-up-d: dev-env
	docker compose -f docker-compose.dev.yml up --build -d

dev-down:
	docker compose -f docker-compose.dev.yml down -v

dev-logs:
	docker compose -f docker-compose.dev.yml logs -f --tail=120

dev-smoke:
	./scripts/local_smoke_test.sh

mock-online:
	curl -fsS "http://127.0.0.1:18080/admin/scenario?state=online"

mock-offline:
	curl -fsS "http://127.0.0.1:18080/admin/scenario?state=offline"

go-test:
	cd addon && go test ./... -race

ui-test:
	cd addon/frontend && npm install && npm run lint && npm run typecheck && npm run build
