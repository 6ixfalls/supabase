FROM golang:1.26-alpine AS secrets

ARG SUPABASE_SECRETS_VERSION=main

RUN CGO_ENABLED=0 GOBIN=/out \
    go install github.com/6ixfalls/supabase/secrets/cmd/supabase-secrets@${SUPABASE_SECRETS_VERSION}

FROM postgrest/postgrest:v14.12

COPY --from=secrets --chmod=755 /out/supabase-secrets /usr/local/bin/supabase-secrets

ENTRYPOINT ["/usr/local/bin/supabase-secrets", "--service", "rest", "--"]
CMD ["postgrest"]
