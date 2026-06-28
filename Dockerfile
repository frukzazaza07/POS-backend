FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o pos-backend ./server.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/pos-backend .
COPY --from=builder /app/.env* ./
EXPOSE 4000
CMD ["./pos-backend"]
