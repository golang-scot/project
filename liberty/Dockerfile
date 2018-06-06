FROM registry.golang.scot/systemd:latest

ENV APP_ENV prod

COPY liberty /usr/local/bin/

EXPOSE 80 443
CMD ["/usr/local/bin/liberty", "serve"]
