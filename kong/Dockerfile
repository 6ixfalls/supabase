FROM kong/kong:3.9.1
COPY ./kong.yml /home/kong/temp.yml
COPY --chmod=777 ./kong-entrypoint.sh /home/kong/kong-entrypoint.sh
ENTRYPOINT ["/bin/sh", "/home/kong/kong-entrypoint.sh"]