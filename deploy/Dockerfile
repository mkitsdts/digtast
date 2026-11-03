FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/digtast ./cmd/digtast

# ---

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/digtast /usr/local/bin/

ENTRYPOINT ["digtast"]
