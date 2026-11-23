FROM golang:1.26-alpine AS secrets

ARG SUPABASE_SECRETS_VERSION=main

RUN CGO_ENABLED=0 GOBIN=/out \
    go install github.com/6ixfalls/supabase/secrets/cmd/supabase-secrets@${SUPABASE_SECRETS_VERSION}

FROM supabase/realtime:v2.148.3

COPY --from=secrets --chmod=755 /out/supabase-secrets /usr/local/bin/supabase-secrets
COPY --chmod=755 docker-entrypoint.sh /usr/local/bin/realtime-entrypoint.sh

ENTRYPOINT ["/usr/local/bin/supabase-secrets", "--service", "realtime", "--", "/usr/local/bin/realtime-entrypoint.sh"]
CMD ["/app/bin/server"]
