FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

# BuildKit cache
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/main ./cmd/api-gateway/main.go

FROM alpine:3.19 AS run

WORKDIR /app

RUN adduser -D appuser

COPY --from=build /app/main ./main

RUN chown appuser:appuser /app/main

USER appuser

EXPOSE 50051

CMD ["./main"]
