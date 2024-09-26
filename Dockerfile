FROM golang:1.22.1-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY . .

RUN go mod download


RUN go build -o main ./cmd/app

FROM alpine:latest

RUN apk --no-cache add ca-certificates bash netcat-openbsd

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/. .
COPY .env .

COPY wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

CMD ["./wait-for-it.sh", "db", "5432", "--", "./main"]