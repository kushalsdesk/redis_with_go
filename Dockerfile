FROM golang:1.24.7-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod  ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o redis-clone ./app

# RUNTIME

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/redis-clone /app/

EXPOSE 6379

USER nobody

CMD ["./redis-clone"]
