FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/digital-labor .

# ---

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/digital-labor /usr/local/bin/

EXPOSE 10086

ENTRYPOINT ["digital-labor"]
