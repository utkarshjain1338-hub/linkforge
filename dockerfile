FROM golang:1.27.3-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/linkforge ./cmd/api

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && adduser -D -H appuser

WORKDIR /app

COPY --from=builder /out/linkforge /app/linkforge

COPY --from=builder /src/web /app/web

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/linkforge"]