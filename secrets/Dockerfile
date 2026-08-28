# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/supabase-secrets ./cmd/supabase-secrets
COPY internal/secrets ./internal/secrets
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /supabase-secrets ./cmd/supabase-secrets

FROM scratch

COPY --from=build /supabase-secrets /supabase-secrets
