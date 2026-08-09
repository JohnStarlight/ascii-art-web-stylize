FROM golang:1.25.5 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM alpine:latest

LABEL maintainer="balevizo, ppetraki, ivogiake" \
    description="ASCII art web server - Go, no external dependencies" \
    version="1.0"

WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/banners ./banners
COPY --from=builder /app/static ./static

EXPOSE 8080
CMD ["./server"]
