FROM golang:1.27-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/mcping-exporter ./cmd/exporter

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/mcping-exporter /usr/local/bin/mcping-exporter

EXPOSE 9090
ENTRYPOINT ["/usr/local/bin/mcping-exporter"]