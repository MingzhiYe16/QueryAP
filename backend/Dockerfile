FROM golang:1.19 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
