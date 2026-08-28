FROM ghcr.io/6ixfalls/supabase-secrets:latest AS secrets

FROM supabase/realtime:v2.102.3

COPY --from=secrets --chmod=755 /supabase-secrets /usr/local/bin/supabase-secrets

ENTRYPOINT ["/usr/local/bin/supabase-secrets", "--service", "realtime", "--", "/usr/bin/tini", "-s", "-g", "--", "/app/run.sh"]
CMD ["/app/bin/server"]
