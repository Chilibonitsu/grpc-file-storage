FROM golang:alpine AS builder

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod tidy

COPY . .
RUN go build -o /build/server ./cmd/server

FROM alpine:latest

WORKDIR /app
COPY --from=builder /build/server /app/
COPY --from=builder /build/.env /app/.env

EXPOSE 8001

ENTRYPOINT ["/app/server"]