ARG PORT=8080
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o login ./cmd

FROM alpine:latest
ENV APP_PORT=$PORT

RUN adduser -D lennie
USER lennie

WORKDIR /app

COPY --from=builder /app/login .
EXPOSE $APP_PORT

CMD ["./login"]