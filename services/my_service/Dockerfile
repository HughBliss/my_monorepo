FROM golang:1.24-alpine AS base

ARG SERVICE_NAME

FROM base AS builder

RUN apk add --no-cache

WORKDIR /monorepo

COPY go.work go.work
COPY go.work.sum go.work.sum
COPY packages packages
COPY services services

WORKDIR /monorepo/services/${SERVICE_NAME}

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /build/app ./cmd/app/main.go

FROM alpine

ARG SERVICE_NAME

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /build/app .
COPY --from=builder /monorepo/services/${SERVICE_NAME}/prod-config.yaml ./config.yaml

CMD ["./app"]