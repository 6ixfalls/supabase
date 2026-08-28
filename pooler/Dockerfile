FROM golang:1.26-alpine AS secrets

ARG SUPABASE_SECRETS_VERSION=main

RUN CGO_ENABLED=0 GOBIN=/out \
    go install github.com/6ixfalls/supabase/secrets/cmd/supabase-secrets@${SUPABASE_SECRETS_VERSION}

FROM supabase/supavisor:2.9.5

COPY --from=secrets --chmod=755 /out/supabase-secrets /usr/local/bin/supabase-secrets

ADD pooler/pooler.exs /etc/pooler/pooler.exs

ENTRYPOINT ["/usr/local/bin/supabase-secrets", "--service", "supavisor", "--", "/usr/bin/tini", "-s", "-g", "--", "/app/limits.sh"]
CMD ["/bin/sh", "-c", "/app/bin/migrate && /app/bin/supavisor eval \"$(cat /etc/pooler/pooler.exs)\" && exec /app/bin/server"]


