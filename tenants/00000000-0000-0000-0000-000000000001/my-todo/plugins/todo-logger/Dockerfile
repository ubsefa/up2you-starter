FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o todo-logger .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates wget
WORKDIR /app
COPY --from=builder /app/todo-logger .
RUN addgroup -S app && adduser -S app -G app
USER app
EXPOSE 8201
CMD ["./todo-logger"]
