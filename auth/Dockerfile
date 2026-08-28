FROM golang:1.26-alpine AS secrets

ARG SUPABASE_SECRETS_VERSION=main

RUN CGO_ENABLED=0 GOBIN=/out \
    go install github.com/6ixfalls/supabase/secrets/cmd/supabase-secrets@${SUPABASE_SECRETS_VERSION}

FROM supabase/gotrue:v2.189.0

COPY --from=secrets --chmod=755 /out/supabase-secrets /usr/local/bin/supabase-secrets

ENTRYPOINT ["/usr/local/bin/supabase-secrets", "--service", "auth", "--"]
CMD ["auth"]
