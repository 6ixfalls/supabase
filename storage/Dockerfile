FROM ghcr.io/6ixfalls/supabase-secrets:latest AS secrets

FROM supabase/storage-api:v1.60.4

COPY --from=secrets --chmod=755 /supabase-secrets /usr/local/bin/supabase-secrets

ENTRYPOINT ["/usr/local/bin/supabase-secrets", "--service", "storage", "--"]
CMD ["node", "dist/start/server.js"]
