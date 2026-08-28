FROM ghcr.io/6ixfalls/supabase-secrets:latest AS secrets

FROM supabase/gotrue:v2.189.0

COPY --from=secrets --chmod=755 /supabase-secrets /usr/local/bin/supabase-secrets

ENTRYPOINT ["/usr/local/bin/supabase-secrets", "--service", "auth", "--"]
CMD ["auth"]
