
migrate:
	docker run -v ./migrations:/migrations migrate/migrate \
    -path=/migrations -database "postgres://std_user:1234@172.17.0.2:5432/greenlight?sslmode=disable" $(direction)

makemigrations:
	docker run  -v ./migrations:/migrations migrate/migrate create -ext=".sql" -dir="./migrations" $(name)