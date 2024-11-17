current_time = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: api/run
api/run:
	@echo 'Running docker compose services...'
	docker compose up

git_version = $(shell git describe --always --tags --long)
linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${git_version}'
.PHONY: api/build
api/build:
	@echo 'Building cmd/api...'
	go build -o ./bin/api -ldflags=${linker_flags} ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/api ./cmd/api

.PHONY: db/migrations/run
db/migrations/run: confirm
	@echo 'Running ${direction} migrations...'
	migrate -path ./migrations -database $(GREENLIGHT_DB_DSN) $(direction)

.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -dir=./migrations -ext=.sql -seq $(name)

.PHONY: audit 
audit:
	@echo 'Tidying and verifying module dependencies...' 
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...