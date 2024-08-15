run:
	go run ./cmd/api -config=./config/local.yaml
startdb:
	docker run --rm -p 5432:5432 -v greenlight_db:/var/lib/postgresql/data --name greenlight_db \
	-e POSTGRES_PASSWORD=postgres -e POSTGRES_USER=postgres -d postgres:15-alpine
