FROM golang:1.26.2-alpine3.23 AS builder

ARG APP_VERSION

RUN apk --no-cache add git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build-app

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# UTC is default TZ

WORKDIR /app

COPY --from=builder /app/bin/app /app/app

USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/app/app"]
