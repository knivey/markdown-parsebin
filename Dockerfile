FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o dave-web ./cmd/dave-web/

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /build/dave-web /usr/local/bin/dave-web

EXPOSE 8080 8081

ENTRYPOINT ["dave-web"]
