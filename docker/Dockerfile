# ==============================
# Build stage
# ==============================
FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/main /app/cmd/web

# ==============================
# Runtime stage
# ==============================
FROM alpine:3.21 AS runtime

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

RUN adduser -D -g appuser appuser

USER appuser

COPY --from=builder /app/main /app/main

EXPOSE 3333

CMD ["/app/main"]
