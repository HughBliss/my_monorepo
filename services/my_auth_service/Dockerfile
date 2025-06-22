FROM golang:1.24-alpine AS base

ARG SERVICE_NAME

FROM base AS builder

RUN apk add --no-cache

WORKDIR /monorepo

COPY packages packages
COPY services services

WORKDIR /monorepo/services/${SERVICE_NAME}

RUN go mod download

RUN go build -o /build/app ./cmd/app/main.go

FROM alpine

ARG SERVICE_NAME

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /build/app .
COPY --from=builder /monorepo/services/${SERVICE_NAME}/prod-config.yaml ./config.yaml
COPY --from=builder /monorepo/services/${SERVICE_NAME}/locales ./locales

CMD ["./app"]