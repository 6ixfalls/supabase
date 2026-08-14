FROM kong/kong:3.9.3

# Writable path for entrypoint-rendered declarative config (/home/kong is read-only).
ENV KONG_DECLARATIVE_CONFIG=/tmp/kong.yml

COPY ./kong.yml /home/kong/temp.yml
COPY --chmod=755 ./kong-entrypoint.sh /home/kong/kong-entrypoint.sh
ENTRYPOINT ["/bin/sh", "/home/kong/kong-entrypoint.sh"]
