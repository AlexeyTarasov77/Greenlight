runserver:
	go run ./cmd/api -config=./config/local.yaml

startdb:
	docker run --rm -p 5432:5432 -v greenlight_db:/var/lib/postgresql/data --name greenlight_db \
	-e POSTGRES_PASSWORD=postgres -e POSTGRES_USER=postgres -d postgres:15-alpine

run:
	make startdb
	make runserver

stopdb:
	docker stop greenlight_db

migrate:
	docker run -v ./migrations:/migrations migrate/migrate \
    -path=/migrations -database "postgres://std_user:1234@docker.for.mac.localhost:5432/greenlight?sslmode=disable" $(direction)

makemigrations:
	docker run  -v ./migrations:/migrations migrate/migrate create -ext=".sql" -dir="./migrations" $(name)

makemigrations-test:
	docker run  -v ./internal/tests/migrations:/migrations migrate/migrate create -ext=".sql" -dir="./migrations" $(name)

migrate-test:
	docker run -v ./internal/tests/migrations:/migrations migrate/migrate \
	-path=/migrations -database "mysql://web_test:web_test@tcp(docker.for.mac.localhost:3306)/snippetbox_test" $(direction) $(flags)