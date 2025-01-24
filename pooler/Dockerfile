FROM supabase/supavisor:2.1.2

ADD ./pooler.exs /etc/pooler/pooler.exs

CMD ["/bin/sh", "-c", "/app/bin/migrate && /app/bin/supavisor eval \"$$(cat /etc/pooler/pooler.exs)\" && /app/bin/server"]