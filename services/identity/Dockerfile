FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

# BuildKit cache
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/main .

FROM alpine:3.19 AS run

WORKDIR /app

RUN adduser -D appuser

COPY --from=build /app/main ./main

USER appuser

RUN chown appuser:appuser /app/main

EXPOSE 50051

CMD ["./main"]
