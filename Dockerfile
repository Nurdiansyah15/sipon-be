# Stage 1: Builder
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate

# Stage 2: Dev tools (for swagger, lint, etc.)
FROM golang:1.25-alpine AS devtools
RUN go install github.com/swaggo/swag/v2/cmd/swag@latest
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Stage 3: App runtime
FROM alpine:3.19 AS app
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/api .
COPY --from=builder /out/migrate .
COPY migrations/ ./migrations/
EXPOSE 8080
CMD ["./api"]

# Stage 4: Migration runner
FROM alpine:3.19 AS migrate
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/migrate .
COPY migrations/ ./migrations/
ENTRYPOINT ["./migrate"]
