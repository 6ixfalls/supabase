FROM ghcr.io/6ixfalls/supabase-secrets:latest AS secrets

FROM supabase/studio:2026.08.03-sha-022b374

COPY --from=secrets --chmod=755 /supabase-secrets /usr/local/bin/supabase-secrets

ENTRYPOINT ["/usr/local/bin/supabase-secrets", "--service", "studio", "--", "docker-entrypoint.sh"]
CMD ["node", "apps/studio/server.js"]
