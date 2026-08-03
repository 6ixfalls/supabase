FROM envoyproxy/envoy:v1.38.0

# Copy envoy configuration files
COPY envoy.yaml \
    cds.template.yaml \
    lds.template.yaml \
    /etc/envoy/

# Copy envoy entrypoint
COPY --chmod=755 docker-entrypoint.sh /docker-entrypoint.sh

ENTRYPOINT ["/bin/sh", "/docker-entrypoint.sh"]