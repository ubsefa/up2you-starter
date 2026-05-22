FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o care-alerts .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates wget
WORKDIR /app
COPY --from=builder /app/care-alerts .
RUN addgroup -S app && adduser -S app -G app
USER app
EXPOSE 8204
CMD ["./care-alerts"]
