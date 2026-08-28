FROM golang:1.26-alpine AS secrets

ARG SUPABASE_SECRETS_VERSION=main

RUN CGO_ENABLED=0 GOBIN=/out \
    go install github.com/6ixfalls/supabase/secrets/cmd/supabase-secrets@${SUPABASE_SECRETS_VERSION}

FROM envoyproxy/envoy:v1.39.0

COPY --from=secrets --chmod=755 /out/supabase-secrets /usr/local/bin/supabase-secrets

# Copy envoy configuration files
COPY envoy.yaml \
    cds.template.yaml \
    lds.template.yaml \
    /etc/envoy/

# Copy envoy entrypoint
COPY --chmod=755 docker-entrypoint.sh /docker-entrypoint.sh

ENTRYPOINT ["/usr/local/bin/supabase-secrets", "--service", "gateway", "--", "/docker-entrypoint.sh"]
