FROM golang:1.22 AS build
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/api/main.go -o internal/api/docs || true

RUN go mod tidy

RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -o /out/migrator ./cmd/migrate

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/api /app/api
COPY --from=build /out/migrator /app/migrator
EXPOSE 8080
ENTRYPOINT ["/app/api"]
