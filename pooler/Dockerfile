FROM ghcr.io/6ixfalls/supabase-secrets:latest AS secrets

FROM supabase/supavisor:2.9.5

COPY --from=secrets --chmod=755 /supabase-secrets /usr/local/bin/supabase-secrets

ADD pooler.exs /etc/pooler/pooler.exs

ENTRYPOINT ["/usr/local/bin/supabase-secrets", "--service", "supavisor", "--", "/usr/bin/tini", "-s", "-g", "--", "/app/limits.sh"]
CMD ["/bin/sh", "-c", "/app/bin/migrate && /app/bin/supavisor eval \"$(cat /etc/pooler/pooler.exs)\" && exec /app/bin/server"]


