FROM ghcr.io/6ixfalls/supabase-secrets:latest AS secrets

FROM envoyproxy/envoy:v1.39.0

COPY --from=secrets --chmod=755 /supabase-secrets /usr/local/bin/supabase-secrets

# Copy envoy configuration files
COPY envoy.yaml \
    cds.template.yaml \
    lds.template.yaml \
    /etc/envoy/

# Copy envoy entrypoint
COPY --chmod=755 docker-entrypoint.sh /docker-entrypoint.sh

ENTRYPOINT ["/usr/local/bin/supabase-secrets", "--service", "gateway", "--", "/docker-entrypoint.sh"]
