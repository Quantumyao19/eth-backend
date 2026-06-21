FROM golang:1.26.2 AS builder

WORKDIR /app
COPY . .

RUN go mod tidy
# Ensure to compile Linux-compatible binary files within the container
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
RUN chmod +x ./main

CMD ["./main"]
