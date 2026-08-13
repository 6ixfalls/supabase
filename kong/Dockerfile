FROM kong/kong:3.9.3

# Writable path for entrypoint-rendered declarative config (/home/kong is read-only).
ENV KONG_DECLARATIVE_CONFIG=/tmp/kong.yml
ENV KONG_PLUGINS="request-transformer,cors,key-auth,acl,basic-auth,request-termination,post-function"
ENV KONG_ROUTER_FLAVOR=expressions
ENV FUNCTIONS_HOST=127.0.0.1

COPY ./kong.yml /home/kong/temp.yml
COPY --chmod=755 ./kong-entrypoint.sh /home/kong/kong-entrypoint.sh
ENTRYPOINT ["/bin/sh", "/home/kong/kong-entrypoint.sh"]
