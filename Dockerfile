FROM golang:tip-alpine3.23 AS builder
WORKDIR /app
COPY . .
RUN go build -o main cmd/api/main.go

FROM alpine:3.23
WORKDIR /app
COPY --from=builder /app/main .
COPY .env .
EXPOSE 8080
CMD ["./main"]