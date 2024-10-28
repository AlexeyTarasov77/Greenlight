FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENTRYPOINT [ "go", "run", "./cmd/api", "-config=./config/local.yaml" ]