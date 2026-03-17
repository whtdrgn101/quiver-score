FROM golang:1.26-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /build/api /usr/local/bin/api
ENTRYPOINT ["api"]
