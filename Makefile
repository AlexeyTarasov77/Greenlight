.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: db/migrations
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