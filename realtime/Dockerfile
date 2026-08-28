FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS secrets

ARG TARGETOS
ARG TARGETARCH
ARG SUPABASE_SECRETS_VERSION=main

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOBIN=/out \
    go install github.com/6ixfalls/supabase/secrets/cmd/supabase-secrets@${SUPABASE_SECRETS_VERSION}

FROM supabase/realtime:v2.102.3

COPY --from=secrets --chmod=755 /out/supabase-secrets /usr/local/bin/supabase-secrets

ENTRYPOINT ["/usr/local/bin/supabase-secrets", "--service", "realtime", "--", "/usr/bin/tini", "-s", "-g", "--", "/app/run.sh"]
CMD ["/app/bin/server"]
