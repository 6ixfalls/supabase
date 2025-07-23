FROM supabase/postgres:15.8.1.121

ADD ./realtime.sql /docker-entrypoint-initdb.d/migrations/99-realtime.sql
ADD ./webhooks.sql /docker-entrypoint-initdb.d/init-scripts/98-webhooks.sql
ADD ./roles.sql /docker-entrypoint-initdb.d/init-scripts/99-roles.sql
ADD ./jwt.sql /docker-entrypoint-initdb.d/init-scripts/99-jwt.sql
ADD ./_supabase.sql /docker-entrypoint-initdb.d/migrations/97-_supabase.sql
ADD ./logs.sql /docker-entrypoint-initdb.d/migrations/99-logs.sql
ADD ./pooler.sql /docker-entrypoint-initdb.d/migrations/99-pooler.sql

COPY --chmod=755 wrapper.sh /usr/local/bin/wrapper.sh

ENTRYPOINT ["wrapper.sh"]
CMD ["postgres", "-c", "config_file=/etc/postgresql/postgresql.conf", "--port=5432"]