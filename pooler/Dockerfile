FROM supabase/supavisor:1.1.56

ADD ./pooler.exs /etc/pooler/pooler.exs

CMD ["/bin/sh", "-c", "/app/bin/migrate && /app/bin/supavisor eval \"$$(cat /etc/pooler/pooler.exs)\" && /app/bin/server"]