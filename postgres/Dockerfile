ARG POSTGRES_VERSION=17.6.1.158

FROM golang:1.26-alpine AS secrets

ARG SUPABASE_SECRETS_VERSION=main

RUN CGO_ENABLED=0 GOBIN=/out \
    go install github.com/6ixfalls/supabase/secrets/cmd/supabase-secrets@${SUPABASE_SECRETS_VERSION}

FROM supabase/postgres:${POSTGRES_VERSION}

COPY --from=secrets --chmod=755 /out/supabase-secrets /usr/local/bin/supabase-secrets

ADD postgres/realtime.sql /docker-entrypoint-initdb.d/migrations/99-realtime.sql
ADD postgres/webhooks.sql /docker-entrypoint-initdb.d/init-scripts/98-webhooks.sql
ADD postgres/roles.sql /docker-entrypoint-initdb.d/init-scripts/99-roles.sql
ADD postgres/jwt.sql /docker-entrypoint-initdb.d/init-scripts/99-jwt.sql
ADD postgres/_supabase.sql /docker-entrypoint-initdb.d/migrations/97-_supabase.sql
ADD postgres/logs.sql /docker-entrypoint-initdb.d/migrations/99-logs.sql
ADD postgres/pooler.sql /docker-entrypoint-initdb.d/migrations/99-pooler.sql

COPY --chmod=755 postgres/wrapper.sh /usr/local/bin/wrapper.sh

ENTRYPOINT ["/usr/local/bin/supabase-secrets", "--service", "db", "--", "/usr/local/bin/wrapper.sh"]

CMD ["postgres", "-c", "config_file=/etc/postgresql/postgresql.conf", "--port=5432"]
