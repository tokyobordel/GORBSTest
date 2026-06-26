FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app .

FROM ubuntu:22.04

WORKDIR /app

COPY --from=builder /app .

COPY --from=builder /build/static ./static

EXPOSE 8080

CMD ["./app"]