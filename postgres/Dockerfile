ARG POSTGRES_VERSION=17.6.1.158

FROM ghcr.io/6ixfalls/supabase-secrets:latest AS secrets

FROM supabase/postgres:${POSTGRES_VERSION}

COPY --from=secrets --chmod=755 /supabase-secrets /usr/local/bin/supabase-secrets

ADD realtime.sql /docker-entrypoint-initdb.d/migrations/99-realtime.sql
ADD webhooks.sql /docker-entrypoint-initdb.d/init-scripts/98-webhooks.sql
ADD roles.sql /docker-entrypoint-initdb.d/init-scripts/99-roles.sql
ADD jwt.sql /docker-entrypoint-initdb.d/init-scripts/99-jwt.sql
ADD _supabase.sql /docker-entrypoint-initdb.d/migrations/97-_supabase.sql
ADD logs.sql /docker-entrypoint-initdb.d/migrations/99-logs.sql
ADD pooler.sql /docker-entrypoint-initdb.d/migrations/99-pooler.sql

COPY --chmod=755 wrapper.sh /usr/local/bin/wrapper.sh

ENTRYPOINT ["/usr/local/bin/supabase-secrets", "--service", "db", "--", "/usr/local/bin/wrapper.sh"]

CMD ["postgres", "-c", "config_file=/etc/postgresql/postgresql.conf", "--port=5432"]
