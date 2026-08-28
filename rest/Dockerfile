FROM ghcr.io/6ixfalls/supabase-secrets:latest AS secrets

FROM postgrest/postgrest:v14.12

COPY --from=secrets --chmod=755 /supabase-secrets /usr/local/bin/supabase-secrets

ENTRYPOINT ["/usr/local/bin/supabase-secrets", "--service", "rest", "--"]
CMD ["postgrest"]
