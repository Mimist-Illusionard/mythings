FROM golang:1.26.1-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/mythings ./cmd/mythings

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/mythings /app/mythings
COPY web /app/web
COPY .env.example /app/.env.example

RUN mkdir -p /app/web/uploads

EXPOSE 8080

ENTRYPOINT ["/app/mythings"]
CMD ["-env=/app/.env.example", "-port=8080"]
